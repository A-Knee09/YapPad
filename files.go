// NOTE: Only for Disk operations and Image rendering logic

package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

var glamourRenderer, _ = glamour.NewTermRenderer(
	glamour.WithAutoStyle(),
	glamour.WithWordWrap(0),
)

/*
	NOTE:

readFile loads text file content with syntax highlighting.
Images are NOT handled here — they use renderImage() instead.
*/
func readFile(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return fileLoadedMsg{content: "Error reading file"}
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".md", ".markdown", ".txt", ".go", ".c", ".cpp", ".h", ".py", ".js", ".ts", ".html", ".css", ".json", ".yaml", ".yml", ".toml", ".sh", ".mod", ".sum":
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

		if ext == ".md" || ext == ".markdown" {
			rendered, err := glamourRenderer.Render(string(content))
			if err != nil {
				return fileLoadedMsg{content: string(content)}
			}
			return fileLoadedMsg{content: rendered}
		}

		var buf bytes.Buffer
		err = quick.Highlight(&buf, string(content), "markdown", "terminal256", "monokai")
		if err != nil {
			return fileLoadedMsg{content: string(content)}
		}
		return fileLoadedMsg{content: buf.String()}
	}
}

// NOTE: Made for adding description to an item
func writeMetaDesc(filePath, desc string) error {
	if desc == "" {
		return nil
	}
	metaDir := filepath.Join(vaultDir, ".metadesc")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return err
	}
	rel, err := filepath.Rel(vaultDir, filePath)
	if err != nil {
		rel = filepath.Base(filePath)
	}
	key := strings.ReplaceAll(rel, string(filepath.Separator), "__")
	metaPath := filepath.Join(metaDir, key+".meta")
	return os.WriteFile(metaPath, []byte(desc), 0o644)
}

func readMetaDesc(filePath string) string {
	metaDir := filepath.Join(vaultDir, ".metadesc")
	rel, err := filepath.Rel(vaultDir, filePath)
	if err != nil {
		rel = filepath.Base(filePath)
	}
	key := strings.ReplaceAll(rel, string(filepath.Separator), "__")
	metaPath := filepath.Join(metaDir, key+".meta")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func deleteMetaDesc(filePath string) {
	metaDir := filepath.Join(vaultDir, ".metadesc")
	rel, err := filepath.Rel(vaultDir, filePath)
	if err != nil {
		rel = filepath.Base(filePath)
	}
	key := strings.ReplaceAll(rel, string(filepath.Separator), "__")
	metaPath := filepath.Join(metaDir, key+".meta")
	os.Remove(metaPath)
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
			creTime = getCreationTime(stat)
		} else {
			creTime = modTime
		}
		modStr := modTime.Format(time.RFC822)

		customDesc := readMetaDesc(path)
		var desc string
		if customDesc != "" {
			desc = customDesc
		} else {
			desc = "Modified: " + modStr
		}

		var displayName string
		if yMode == yapAll {
			// Show relative path from vault root (includes subdir prefix)
			displayName, _ = filepath.Rel(vaultDir, path)
		} else {
			// Show just the filename within the subdir
			displayName = d.Name()
		}

		items = append(items, item{
			title: displayName,
			desc:  desc,
			// desc:    "Modified: " + modStr,
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
		case sortNameDesc:
			return strings.ToLower(itemI.title) > strings.ToLower(itemJ.title)
		case sortNameAsc:
			return strings.ToLower(itemI.title) < strings.ToLower(itemJ.title)
		default:
			return itemI.modTime.After(itemJ.modTime)
		}
	})

	return items
}
