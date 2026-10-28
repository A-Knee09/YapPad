// NOTE: All image rendering logic here :D
package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sys/unix"
)

// isImageFile checks if a file is an image by extension or content detection.
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".svg", ".ico", ".tiff":
		return true
	}
	// Fallback: read first 512 bytes and detect
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if n == 0 {
		return false
	}
	ct := http.DetectContentType(buf[:n])
	return strings.HasPrefix(ct, "image/")
}

// clearKittyGraphics sends the escape sequence to delete all Kitty images.
func clearKittyGraphics() tea.Cmd {
	requestID := invalidateImageRequest()
	return func() tea.Msg {
		writeImageIfCurrent(requestID, []byte("\x1b_Ga=d,d=a,q=1\x1b\\"))
		return nil
	}
}

// HACK: Getting ascpect ratio helps me to draw 9:16,4:5 images away from list, with better height. Without this it looked like the edge of list and edge of image collapsed together
func getImageAspectRatio(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 1.0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 1.0, err
	}

	if cfg.Width == 0 {
		return 1.0, nil
	}
	return float64(cfg.Height) / float64(cfg.Width), nil
}

// Kitty terminals receive a cached image asset once, then resize by replacing
// its placement. Other environments keep using Chafa's Kitty output.
var (
	imageCache          = map[string][]byte{}
	imageCacheOrder     []string
	imageCacheBytes     int
	imageCacheMu        sync.Mutex
	imageOutputMu       sync.Mutex
	imageRequestMu      sync.Mutex
	imageRequestID      uint64
	imageCancelFunc     context.CancelFunc
	kittyAssets         = map[string]kittyImageAsset{}
	kittyAssetOrder     []string
	kittyPendingDeletes []uint32
	kittyAssetsBytes    int64
	kittyAssetsMu       sync.Mutex
	kittyImageSequence  uint32
)

const imageCacheLimitBytes = 64 << 20
const kittyAssetCacheLimitBytes = 96 << 20

type kittyImageAsset struct {
	key      string
	path     string
	id       uint32
	tempPath string
	uploaded bool
	size     int64
	width    int
	height   int
}

func newImageRequest() (uint64, context.Context) {
	imageRequestMu.Lock()
	defer imageRequestMu.Unlock()
	if imageCancelFunc != nil {
		imageCancelFunc()
	}
	imageRequestID++
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	imageCancelFunc = cancel
	return imageRequestID, ctx
}

func invalidateImageRequest() uint64 {
	imageRequestMu.Lock()
	defer imageRequestMu.Unlock()
	if imageCancelFunc != nil {
		imageCancelFunc()
		imageCancelFunc = nil
	}
	imageRequestID++
	return imageRequestID
}

func isCurrentImageRequest(requestID uint64) bool {
	imageRequestMu.Lock()
	defer imageRequestMu.Unlock()
	return requestID == imageRequestID
}

func writeImageIfCurrent(requestID uint64, content []byte) bool {
	imageRequestMu.Lock()
	defer imageRequestMu.Unlock()
	if requestID != imageRequestID {
		return false
	}
	imageOutputMu.Lock()
	defer imageOutputMu.Unlock()
	_, _ = os.Stdout.Write(content)
	return true
}

func kittyGraphicsAvailable() bool {
	if os.Getenv("SSH_CONNECTION") != "" {
		return false
	}
	return os.Getenv("KITTY_WINDOW_ID") != "" || os.Getenv("TERM") == "xterm-kitty"
}

func moveKeyToEnd(order []string, key string) []string {
	for i, entry := range order {
		if entry == key {
			order = append(order[:i], order[i+1:]...)
			break
		}
	}
	return append(order, key)
}

func getKittyImageAsset(ctx context.Context, path string) (kittyImageAsset, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return kittyImageAsset{}, false
	}
	key := fmt.Sprintf("%s:%d:%d", path, info.Size(), info.ModTime().UnixNano())

	kittyAssetsMu.Lock()
	if asset, ok := kittyAssets[key]; ok {
		kittyAssetOrder = moveKeyToEnd(kittyAssetOrder, key)
		kittyAssetsMu.Unlock()
		return asset, true
	}
	kittyAssetsMu.Unlock()

	probe, err := os.Open(path)
	if err != nil {
		return kittyImageAsset{}, false
	}
	config, format, configErr := image.DecodeConfig(probe)
	if configErr != nil || ctx.Err() != nil || int64(config.Width)*int64(config.Height)*4 > kittyAssetCacheLimitBytes {
		_ = probe.Close()
		return kittyImageAsset{}, false
	}

	assetPath := path
	var tempPath string
	if format != "png" {
		_, err = probe.Seek(0, 0)
		if err != nil {
			_ = probe.Close()
			return kittyImageAsset{}, false
		}
		decoded, _, decodeErr := image.Decode(probe)
		_ = probe.Close()
		if decodeErr != nil || ctx.Err() != nil {
			return kittyImageAsset{}, false
		}
		temp, err := os.CreateTemp("", "yappad-image-*.png")
		if err != nil {
			return kittyImageAsset{}, false
		}
		tempPath = temp.Name()
		encoder := png.Encoder{CompressionLevel: png.NoCompression}
		encodeErr := encoder.Encode(temp, decoded)
		closeErr := temp.Close()
		if encodeErr != nil || closeErr != nil || ctx.Err() != nil {
			_ = os.Remove(tempPath)
			return kittyImageAsset{}, false
		}
		assetPath = tempPath
	} else {
		_ = probe.Close()
	}
	assetInfo, err := os.Stat(assetPath)
	if err != nil || assetInfo.Size() > kittyAssetCacheLimitBytes {
		if tempPath != "" {
			_ = os.Remove(tempPath)
		}
		return kittyImageAsset{}, false
	}

	kittyAssetsMu.Lock()
	if existing, ok := kittyAssets[key]; ok {
		kittyAssetOrder = moveKeyToEnd(kittyAssetOrder, key)
		kittyAssetsMu.Unlock()
		if tempPath != "" {
			_ = os.Remove(tempPath)
		}
		return existing, true
	}
	if kittyImageSequence == 0 {
		kittyImageSequence = uint32(1_000_000_000 + (os.Getpid()%1_000_000)*1_000)
	}
	kittyImageSequence++
	asset := kittyImageAsset{
		key: key, path: assetPath, tempPath: tempPath, id: kittyImageSequence,
		size: assetInfo.Size(), width: config.Width, height: config.Height,
	}
	kittyAssets[key] = asset
	kittyAssetOrder = append(kittyAssetOrder, key)
	kittyAssetsBytes += asset.size
	for len(kittyAssetOrder) > 12 || kittyAssetsBytes > kittyAssetCacheLimitBytes {
		oldest := kittyAssetOrder[0]
		kittyAssetOrder = kittyAssetOrder[1:]
		oldAsset := kittyAssets[oldest]
		delete(kittyAssets, oldest)
		kittyAssetsBytes -= oldAsset.size
		kittyPendingDeletes = append(kittyPendingDeletes, oldAsset.id)
		if oldAsset.tempPath != "" {
			_ = os.Remove(oldAsset.tempPath)
		}
	}
	kittyAssetsMu.Unlock()
	return asset, true
}

func renderKittyImage(requestID uint64, asset kittyImageAsset, cols, rows, xOffset, yOffset int) tea.Msg {
	imageRequestMu.Lock()
	defer imageRequestMu.Unlock()
	if requestID != imageRequestID {
		return nil
	}
	imageOutputMu.Lock()
	defer imageOutputMu.Unlock()

	kittyAssetsMu.Lock()
	currentAsset, ok := kittyAssets[asset.key]
	if !ok {
		kittyAssetsMu.Unlock()
		return nil
	}
	var buf bytes.Buffer
	for _, id := range kittyPendingDeletes {
		buf.WriteString(fmt.Sprintf("\x1b_Ga=d,d=i,i=%d,q=1\x1b\\", id))
	}
	kittyPendingDeletes = nil
	fitCols, fitRows := fitImageToCells(cols, rows, asset.width, asset.height)
	xOffset += max(0, (cols-fitCols)/2)
	yOffset += max(0, (rows-fitRows)/2)
	buf.WriteString("\x1b[s")
	buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", yOffset, xOffset))
	if currentAsset.uploaded {
		buf.WriteString(fmt.Sprintf("\x1b_Ga=p,i=%d,p=1,c=%d,r=%d,C=1,q=1\x1b\\", asset.id, fitCols, fitRows))
	} else {
		payload := base64.StdEncoding.EncodeToString([]byte(asset.path))
		buf.WriteString(fmt.Sprintf("\x1b_Ga=T,t=f,f=100,i=%d,p=1,c=%d,r=%d,C=1,q=1;%s\x1b\\", asset.id, fitCols, fitRows, payload))
		currentAsset.uploaded = true
		kittyAssets[asset.key] = currentAsset
	}
	buf.WriteString("\x1b[u")
	_, _ = os.Stdout.Write(buf.Bytes())
	kittyAssetsMu.Unlock()
	return imageRenderedMsg{requestID: requestID}
}

// fitImageToCells preserves the image's pixel aspect ratio while accounting
// for terminal cells being roughly twice as tall as they are wide.
func fitImageToCells(maxCols, maxRows, imageWidth, imageHeight int) (int, int) {
	if maxCols < 1 || maxRows < 1 || imageWidth < 1 || imageHeight < 1 {
		return max(1, maxCols), max(1, maxRows)
	}
	cellAspect := terminalCellAspectRatio()
	colsPerRow := (float64(imageWidth) / float64(imageHeight)) / cellAspect
	cols, rows := maxCols, int(float64(maxCols)/colsPerRow+0.5)
	if rows > maxRows {
		rows = maxRows
		cols = int(float64(maxRows)*colsPerRow + 0.5)
	}
	return min(maxCols, max(1, cols)), min(maxRows, max(1, rows))
}

func terminalCellAspectRatio() float64 {
	window, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err == nil && window.Col > 0 && window.Row > 0 && window.Xpixel > 0 && window.Ypixel > 0 {
		cellWidth := float64(window.Xpixel) / float64(window.Col)
		cellHeight := float64(window.Ypixel) / float64(window.Row)
		if cellHeight > 0 {
			return cellWidth / cellHeight
		}
	}
	// Common terminal cells are about twice as tall as they are wide.
	return 0.5
}

func cleanupKittyImageCache() {
	imageRequestMu.Lock()
	if imageCancelFunc != nil {
		imageCancelFunc()
		imageCancelFunc = nil
	}
	imageRequestID++
	imageOutputMu.Lock()
	kittyAssetsMu.Lock()
	for _, asset := range kittyAssets {
		_, _ = fmt.Fprintf(os.Stdout, "\x1b_Ga=d,d=i,i=%d,q=1\x1b\\", asset.id)
		if asset.tempPath != "" {
			_ = os.Remove(asset.tempPath)
		}
	}
	for _, id := range kittyPendingDeletes {
		_, _ = fmt.Fprintf(os.Stdout, "\x1b_Ga=d,d=i,i=%d,q=1\x1b\\", id)
	}
	kittyAssets = map[string]kittyImageAsset{}
	kittyAssetOrder = nil
	kittyPendingDeletes = nil
	kittyAssetsBytes = 0
	kittyAssetsMu.Unlock()
	imageOutputMu.Unlock()
	imageRequestMu.Unlock()
}

func getCachedChafaImage(key string) ([]byte, bool) {
	imageCacheMu.Lock()
	defer imageCacheMu.Unlock()
	content, ok := imageCache[key]
	if ok {
		imageCacheOrder = moveKeyToEnd(imageCacheOrder, key)
	}
	return content, ok
}

func cacheChafaImage(key string, content []byte) {
	imageCacheMu.Lock()
	defer imageCacheMu.Unlock()
	if previous, ok := imageCache[key]; ok {
		imageCacheBytes -= len(previous)
	} else {
		imageCacheOrder = append(imageCacheOrder, key)
	}
	imageCache[key] = content
	imageCacheBytes += len(content)
	imageCacheOrder = moveKeyToEnd(imageCacheOrder, key)
	for imageCacheBytes > imageCacheLimitBytes && len(imageCacheOrder) > 0 {
		oldest := imageCacheOrder[0]
		imageCacheOrder = imageCacheOrder[1:]
		if old, ok := imageCache[oldest]; ok {
			imageCacheBytes -= len(old)
			delete(imageCache, oldest)
		}
	}
}

// quietKittyGraphics suppresses success replies for every Kitty command.
// Chafa can omit the quiet flag on continuation chunks, allowing terminal
// acknowledgements to leak into Bubble Tea as keyboard input.
func quietKittyGraphics(content []byte) []byte {
	start, end := []byte("\x1b_G"), []byte("\x1b\\")
	var out bytes.Buffer
	for offset := 0; offset < len(content); {
		relStart := bytes.Index(content[offset:], start)
		if relStart < 0 {
			out.Write(content[offset:])
			break
		}
		cmdStart := offset + relStart
		out.Write(content[offset:cmdStart])
		relEnd := bytes.Index(content[cmdStart+len(start):], end)
		if relEnd < 0 {
			out.Write(content[cmdStart:])
			break
		}
		cmdEnd := cmdStart + len(start) + relEnd
		command := content[cmdStart+len(start) : cmdEnd]
		headerEnd := bytes.IndexByte(command, ';')
		payload := []byte(nil)
		if headerEnd >= 0 {
			payload = command[headerEnd:]
		} else {
			headerEnd = len(command)
		}
		header := bytes.Split(command[:headerEnd], []byte{','})
		quietSet := false
		for i, field := range header {
			if bytes.HasPrefix(field, []byte("q=")) {
				header[i] = []byte("q=1")
				quietSet = true
			}
		}
		if !quietSet {
			header = append(header, []byte("q=1"))
		}
		out.Write(start)
		out.Write(bytes.Join(header, []byte{','}))
		out.Write(payload)
		out.Write(end)
		offset = cmdEnd + len(end)
	}
	return out.Bytes()
}

func renderImage(path string, cols, rows, xOffset, yOffset int, requestID uint64, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		if kittyGraphicsAvailable() {
			if asset, ok := getKittyImageAsset(ctx, path); ok {
				return renderKittyImage(requestID, asset, cols, rows, xOffset, yOffset)
			}
		}

		// Remove the previous Kitty image before doing any potentially slow work.
		// This matters during terminal resizes, where the old image can otherwise
		// cover the loading indicator until chafa finishes.
		if !writeImageIfCurrent(requestID, []byte("\x1b_Ga=d,d=a,q=1\x1b\\")) {
			return nil
		}

		key := fmt.Sprintf("%s-%dx%d", path, cols, rows)

		cached, ok := getCachedChafaImage(key)

		if ok {
			var buf bytes.Buffer
			buf.WriteString("\x1b[s")
			buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", yOffset, xOffset))
			buf.Write(cached)
			buf.WriteString("\x1b[u")
			if !writeImageIfCurrent(requestID, buf.Bytes()) {
				return nil
			}
			return imageRenderedMsg{requestID: requestID}
		}

		cmd := exec.CommandContext(ctx, "chafa", "-f", "kitty", "-s", fmt.Sprintf("%dx%d", cols, rows), path)
		output, err := cmd.Output()
		if err != nil {
			imageRequestMu.Lock()
			current := requestID == imageRequestID
			imageRequestMu.Unlock()
			if !current {
				return nil
			}
			return imageRenderedMsg{requestID: requestID, err: err}
		}
		output = quietKittyGraphics(output)

		cacheChafaImage(key, output)

		var buf bytes.Buffer
		buf.WriteString("\x1b[s")
		buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", yOffset, xOffset))
		buf.Write(output)
		buf.WriteString("\x1b[u")
		if !writeImageIfCurrent(requestID, buf.Bytes()) {
			return nil
		}
		return imageRenderedMsg{requestID: requestID}
	}
}

func openImageViewer(path string) tea.Cmd {
	viewer := cmp.Or(os.Getenv("IMAGE_VIEWER"), "xdg-open")
	c := exec.Command(viewer, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return fileEditedMsg{err: err}
	})
}
