package assets

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyAssets(sourceDir, assetsDir, destDir string) (int, error) {
	sourcePath := filepath.Join(sourceDir, assetsDir)
	destPath := filepath.Join(destDir, assetsDir)

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return 0, nil
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return 0, err
	}

	count := 0

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destFile := filepath.Join(destPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(destFile, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open asset %s: %w", path, err)
		}
		defer src.Close()

		dest, err := os.Create(destFile)
		if err != nil {
			return fmt.Errorf("failed to create asset %s: %w", destFile, err)
		}
		defer dest.Close()

		if _, err := io.Copy(dest, src); err != nil {
			return fmt.Errorf("failed to copy asset %s: %w", path, err)
		}

		count++
		return nil
	})

	return count, err
}
