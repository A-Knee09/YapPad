// NOTE: All image rendering logic here :D
package main

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
		writeImageIfCurrent(requestID, []byte("\x1b_Ga=d,d=a\x1b\\"))
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

/*
	NOTE:

renderImage uses chafa to generate a Kitty image escape sequence, then
writes it directly to stdout at a specific cell offset. The output is
captured first to prevent chafa's own cursor movements from wrecking
the Bubble Tea TUI.
*/
var (
	imageCache      = map[string][]byte{}
	imageCacheMu    sync.Mutex
	imageOutputMu   sync.Mutex
	imageRequestMu  sync.Mutex
	imageRequestID  uint64
	imageCancelFunc context.CancelFunc
)

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

func renderImage(path string, cols, rows, xOffset, yOffset int, requestID uint64, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("%s-%dx%d", path, cols, rows)

		imageCacheMu.Lock()
		cached, ok := imageCache[key]
		imageCacheMu.Unlock()

		if ok {
			var buf bytes.Buffer
			buf.WriteString("\x1b_Ga=d,d=a\x1b\\")
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

		imageCacheMu.Lock()
		imageCache[key] = output
		imageCacheMu.Unlock()

		var buf bytes.Buffer
		buf.WriteString("\x1b_Ga=d,d=a\x1b\\")
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
