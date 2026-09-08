package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyAssets(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	assetsDir := "assets"

	// Create source directory structure
	if err := os.MkdirAll(filepath.Join(sourceDir, assetsDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create some test files
	files := map[string]string{
		"file1.txt":           "content1",
		"subdir/file2.txt":    "content2",
		"subdir/nested/f.txt": "nested content",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(sourceDir, assetsDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", relPath, err)
		}
	}

	// Run CopyAssets
	count, err := CopyAssets(sourceDir, assetsDir, destDir)
	if err != nil {
		t.Fatalf("CopyAssets failed: %v", err)
	}

	if count != len(files) {
		t.Errorf("Expected to copy %d files, got %d", len(files), count)
	}

	// Verify destination structure
	for relPath, expectedContent := range files {
		destFile := filepath.Join(destDir, assetsDir, relPath)
		content, err := os.ReadFile(destFile)
		if err != nil {
			t.Errorf("Failed to read destination file %s: %v", destFile, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("File %s: expected content %q, got %q", relPath, expectedContent, string(content))
		}
	}
}

func TestCopyAssets_NoSource(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")

	// Source directory doesn't exist
	count, err := CopyAssets(sourceDir, "nonexistent", destDir)
	if err != nil {
		t.Errorf("CopyAssets should not error when source doesn't exist: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Destination should not be created
	if _, err := os.Stat(filepath.Join(destDir, "nonexistent")); !os.IsNotExist(err) {
		t.Error("Destination directory should not exist")
	}
}

func TestCopyAssets_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	assetsDir := "empty"

	// Create empty assets directory
	if err := os.MkdirAll(filepath.Join(sourceDir, assetsDir), 0755); err != nil {
		t.Fatalf("Failed to create empty source dir: %v", err)
	}

	count, err := CopyAssets(sourceDir, assetsDir, destDir)
	if err != nil {
		t.Fatalf("CopyAssets failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0 for empty dir, got %d", count)
	}
}
