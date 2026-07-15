package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFoldersAndTree(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Dynamically create test vault structure:
	// tempDir/file1.md
	// tempDir/file2.txt
	// tempDir/sub1/note3.md
	// tempDir/sub1/sub2/note4.md

	err := os.WriteFile(filepath.Join(tempDir, "file1.md"), []byte("content"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(filepath.Join(tempDir, "sub1", "sub2"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "sub1", "note3.md"), []byte("content"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "sub1", "sub2", "note4.md"), []byte("content"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test buildTree
	tree := buildTree(tempDir, "")
	t.Logf("Generated tree:\n%s", tree)

	if !strings.Contains(tree, "file1.md") {
		t.Errorf("Expected tree to contain file1.md")
	}
	if !strings.Contains(tree, "sub1") {
		t.Errorf("Expected tree to contain sub1")
	}
	if !strings.Contains(tree, "note3.md") {
		t.Errorf("Expected tree to contain note3.md")
	}

	// Test listFolders
	vaultDir = tempDir // Set global vaultDir to tempDir
	folders := listFolders(tempDir)
	t.Logf("List folders:")
	for _, it := range folders {
		item := it.(item)
		t.Logf("- Title: %s, isDir: %v, isRoot: %v, path: %s", item.title, item.isDir, item.isRoot, item.path)
	}

	if len(folders) != 2 {
		t.Errorf("Expected 2 items in folders list, got %d", len(folders))
	}
	if folders[0].(item).title != "Files in this folder" {
		t.Errorf("Expected first item to be root option")
	}
	if folders[1].(item).title != "sub1" {
		t.Errorf("Expected second item to be sub1 directory")
	}

	// Test listFilesOnly
	files := listFilesOnly(tempDir, sortNameAsc)
	t.Logf("List files:")
	for _, it := range files {
		item := it.(item)
		t.Logf("- Title: %s, path: %s", item.title, item.path)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files in root, got %d", len(files))
	}
	if files[0].(item).title != "file1.md" || files[1].(item).title != "file2.txt" {
		t.Errorf("Expected file1.md and file2.txt, got %s and %s", files[0].(item).title, files[1].(item).title)
	}
}
