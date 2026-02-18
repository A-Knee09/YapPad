package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/list"
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
	return func() tea.Msg {
		fmt.Print("\x1b_Ga=d,d=a\x1b\\")
		return nil
	}
}

// renderImage uses chafa to generate a Kitty image escape sequence, then
// writes it directly to stdout at a specific cell offset. The output is
// *captured* first to prevent chafa's own cursor movements from wrecking
// the Bubble Tea TUI.
func renderImage(path string, cols, rows, xOffset, yOffset int) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("chafa", "-f", "kitty", "-s", fmt.Sprintf("%dx%d", cols, rows), path)
		output, err := cmd.Output()
		if err != nil {
			return imageRenderedMsg{}
		}
		// Save cursor, move to preview pane origin, write image, restore cursor.
		// All written as a single os.Stdout.Write to minimize interference
		// with Bubble Tea's rendering loop.
		var buf bytes.Buffer
		buf.WriteString("\x1b[s")                                     // save cursor
		buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", yOffset, xOffset)) // move to preview pane
		buf.Write(output)                                             // kitty escape sequence
		buf.WriteString("\x1b[u")                                     // restore cursor
		os.Stdout.Write(buf.Bytes())
		return imageRenderedMsg{}
	}
}

// readFile loads text file content with syntax highlighting.
// Images are NOT handled here — they use renderImage() instead.
func readFile(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return fileLoadedMsg{content: "Error reading file"}
		}

		// Check extension first to bypass flaky http.DetectContentType
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".md", ".markdown", ".txt", ".go", ".c", ".cpp", ".h", ".py", ".js", ".ts", ".html", ".css", ".json", ".yaml", ".yml", ".toml", ".sh", ".mod", ".sum":
			// Known text — skip binary detection
		default:
			buffer := make([]byte, 512)
			copy(buffer, content)
			contentType := http.DetectContentType(buffer)

			if strings.HasPrefix(contentType, "audio/") ||
				strings.HasPrefix(contentType, "video/") ||
				contentType == "application/octet-stream" {
				return fileLoadedMsg{content: fmt.Sprintf("[Binary file: %s]", contentType)}
			}
		}

		var buf bytes.Buffer
		err = quick.Highlight(&buf, string(content), "markdown", "terminal256", "monokai")
		if err != nil {
			return fileLoadedMsg{content: string(content)}
		}

		return fileLoadedMsg{content: buf.String()}
	}
}

func getEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "nvim"
}

func openInEditor(path string) tea.Cmd {
	cmd := exec.Command(getEditor(), path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return fileEditedMsg{err: err}
	})
}

func listFiles(sMode sortMode, yMode yapMode) []list.Item {
	var items []list.Item

	var searchDir string
	if yMode == yapAll {
		searchDir = vaultDir
	} else {
		searchDir = filepath.Join(vaultDir, yMode.subdir())
	}

	filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files/directories (starting with .) but NOT the search root
		if d.Name()[0] == '.' && path != searchDir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		modTime := info.ModTime()
		var creTime time.Time

		// Attempt to get creation time (best effort)
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			sec := stat.Ctim.Sec
			nsec := stat.Ctim.Nsec
			creTime = time.Unix(sec, nsec)
		} else {
			creTime = modTime
		}

		modStr := modTime.Format(time.RFC822)

		var displayName string
		if yMode == yapAll {
			// Show relative path from vault root (includes subdir prefix)
			displayName, _ = filepath.Rel(vaultDir, path)
		} else {
			// Show just the filename within the subdir
			displayName = d.Name()
		}

		items = append(items, item{
			title:   displayName,
			desc:    "Modified: " + modStr,
			modTime: modTime,
			creTime: creTime,
		})
		return nil
	})

	sort.Slice(items, func(i, j int) bool {
		itemI := items[i].(item)
		itemJ := items[j].(item)

		switch sMode {
		case sortModifiedDesc:
			return itemI.modTime.After(itemJ.modTime)
		case sortModifiedAsc:
			return itemI.modTime.Before(itemJ.modTime)
		case sortCreatedDesc:
			return itemI.creTime.After(itemJ.creTime)
		case sortCreatedAsc:
			return itemI.creTime.Before(itemJ.creTime)
		default:
			return itemI.modTime.After(itemJ.modTime)
		}
	})

	return items
}
