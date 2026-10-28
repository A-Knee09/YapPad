// NOTE: Only for Disk operations

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
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
)

var (
	glamourRenderer *glamour.TermRenderer
	glamourMu       sync.Mutex
)

func init() {
	style := styles.DarkStyleConfig
	margin := uint(0)
	style.Document.Margin = &margin

	glamourRenderer, _ = glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(0),
	)
}

func renderMarkdown(content string) string {
	glamourMu.Lock()
	defer glamourMu.Unlock()
	rendered, err := glamourRenderer.Render(content)
	if err != nil {
		return content
	}
	return rendered
}

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
			return fileLoadedMsg{content: renderMarkdown(string(content))}
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

func listFiles(sMode sortMode) []list.Item {
	var items []list.Item

	var searchDir string
	searchDir = vaultDir

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

		displayName, _ := filepath.Rel(vaultDir, path)

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

func countFilesRecursively(dirPath string) int {
	count := 0
	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden
		if d.Name()[0] == '.' && path != dirPath {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func listFolders(dir string) []list.Item {
	var items []list.Item

	// 1. Files in this folder
	items = append(items, item{
		title:  "Files in this folder",
		desc:   "View files directly in " + filepath.Base(dir),
		isDir:  true,
		isRoot: true,
		path:   dir,
	})

	entries, err := os.ReadDir(dir)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		if entry.Name()[0] == '.' {
			continue
		}
		if entry.IsDir() {
			path := filepath.Join(dir, entry.Name())
			fileCount := countFilesRecursively(path)
			items = append(items, item{
				title:     entry.Name(),
				desc:      fmt.Sprintf("Folder (%d files)", fileCount),
				isDir:     true,
				path:      path,
				fileCount: fileCount,
			})
		}
	}

	// Sort subdirectories alphabetically (excluding the "Files in this folder" option at index 0)
	if len(items) > 1 {
		sort.Slice(items[1:], func(i, j int) bool {
			return strings.ToLower(items[1+i].(item).title) < strings.ToLower(items[1+j].(item).title)
		})
	}

	return items
}

// listFolderContents presents folders and files together in the current directory.
func listFolderContents(dir string, sMode sortMode) []list.Item {
	var items []list.Item
	entries, err := os.ReadDir(dir)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			items = append(items, item{
				title: entry.Name(), desc: fmt.Sprintf("Folder (%d files)", countFilesRecursively(path)),
				isDir: true, path: path,
			})
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		modTime := info.ModTime()
		creTime := modTime
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			creTime = getCreationTime(stat)
		}
		desc := readMetaDesc(path)
		if desc == "" {
			desc = "Modified: " + modTime.Format(time.RFC822)
		}
		items = append(items, item{title: entry.Name(), desc: desc, path: path, modTime: modTime, creTime: creTime})
	}

	sort.Slice(items, func(i, j int) bool {
		a, b := items[i].(item), items[j].(item)
		if a.isDir != b.isDir {
			return a.isDir
		}
		if a.isDir || sMode == sortNameAsc || sMode == sortNameDesc {
			if sMode == sortNameDesc {
				return strings.ToLower(a.title) > strings.ToLower(b.title)
			}
			return strings.ToLower(a.title) < strings.ToLower(b.title)
		}
		switch sMode {
		case sortModifiedAsc:
			return a.modTime.Before(b.modTime)
		case sortCreatedDesc:
			return a.creTime.After(b.creTime)
		case sortCreatedAsc:
			return a.creTime.Before(b.creTime)
		default:
			return a.modTime.After(b.modTime)
		}
	})
	return items
}

func listFilesOnly(dir string, sMode sortMode) []list.Item {
	var items []list.Item
	entries, err := os.ReadDir(dir)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		if entry.Name()[0] == '.' || entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime()
		var creTime time.Time
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

		items = append(items, item{
			title:   entry.Name(),
			desc:    desc,
			modTime: modTime,
			creTime: creTime,
			path:    path,
		})
	}

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

func buildTree(dirPath string, currentPrefix string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ""
	}

	var validEntries []os.DirEntry
	for _, entry := range entries {
		if entry.Name()[0] == '.' {
			continue
		}
		validEntries = append(validEntries, entry)
	}

	var sb strings.Builder
	for i, entry := range validEntries {
		isLast := i == len(validEntries)-1
		var branch string
		if isLast {
			branch = "└── "
		} else {
			branch = "├── "
		}
		sb.WriteString(currentPrefix + branch + entry.Name() + "\n")
		if entry.IsDir() {
			var nextPrefix string
			if isLast {
				nextPrefix = currentPrefix + "    "
			} else {
				nextPrefix = currentPrefix + "│   "
			}
			sb.WriteString(buildTree(filepath.Join(dirPath, entry.Name()), nextPrefix))
		}
	}
	return sb.String()
}

func folderTreePreview(dirPath string) string {
	tree := buildTree(dirPath, "")
	if tree == "" {
		return "(Empty Folder)\n"
	}
	return tree
}

func pathIsWithinVault(path string) bool {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		return false
	}
	candidate, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func resolveFolderInput(base, input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" || input == "." {
		return base, nil
	}
	dest := input
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(base, dest)
	}
	dest = filepath.Clean(dest)
	if !pathIsWithinVault(dest) {
		return "", fmt.Errorf("destination must stay inside the vault")
	}
	return dest, nil
}

func movePathIntoFolder(source, destFolder string) (string, error) {
	source = filepath.Clean(source)
	destFolder = filepath.Clean(destFolder)
	if !pathIsWithinVault(source) || !pathIsWithinVault(destFolder) {
		return "", fmt.Errorf("source and destination must stay inside the vault")
	}
	info, err := os.Lstat(source)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		rel, err := filepath.Rel(source, destFolder)
		if err != nil {
			return "", err
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			return "", fmt.Errorf("cannot move a folder into itself")
		}
	}
	target := filepath.Join(destFolder, filepath.Base(source))
	if filepath.Clean(target) == source {
		return "", fmt.Errorf("source is already in that folder")
	}
	if _, err := os.Lstat(target); err == nil {
		return "", fmt.Errorf("a file or folder named %q already exists there", filepath.Base(source))
	} else if !os.IsNotExist(err) {
		return "", err
	}

	type metaEntry struct {
		path string
		desc string
	}
	var metadata []metaEntry
	collect := func(path string) {
		if desc := readMetaDesc(path); desc != "" {
			metadata = append(metadata, metaEntry{path: path, desc: desc})
		}
	}
	if info.IsDir() {
		err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() {
				collect(path)
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		collect(source)
	}

	if err := os.MkdirAll(destFolder, 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(source, target); err != nil {
		return "", err
	}
	for _, meta := range metadata {
		rel, err := filepath.Rel(source, meta.path)
		if err != nil {
			continue
		}
		newPath := target
		if rel != "." {
			newPath = filepath.Join(target, rel)
		}
		deleteMetaDesc(meta.path)
		_ = writeMetaDesc(newPath, meta.desc)
	}
	return target, nil
}

func removePathWithMetadata(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		err = filepath.WalkDir(path, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() {
				deleteMetaDesc(filePath)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		deleteMetaDesc(path)
	}
	return os.RemoveAll(path)
}
