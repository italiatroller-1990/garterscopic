package server

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher  *fsnotify.Watcher
	dir      string
	callback func([]string)
	stop     chan struct{}
}

func NewWatcher(dir string) *Watcher {
	return &Watcher{
		dir:  dir,
		stop: make(chan struct{}),
	}
}

// Watch starts watching the source directory. The callback receives the
// (deduplicated) list of files that changed since the last invocation.
func (w *Watcher) Watch(callback func([]string)) {
	w.callback = callback

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}
	w.watcher = watcher

	if err := filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == ".tmp" || name == "dist" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		log.Printf("Failed to add watch: %v", err)
		return
	}

	go w.run()
}

func (w *Watcher) run() {
	debounce := time.NewTicker(100 * time.Millisecond)
	defer debounce.Stop()

	changedFiles := make(map[string]bool)

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename {

				if w.isRelevantFile(event.Name) {
					changedFiles[event.Name] = true
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watch error: %v", err)

		case <-debounce.C:
			if len(changedFiles) > 0 {
				files := make([]string, 0, len(changedFiles))
				for f := range changedFiles {
					files = append(files, f)
				}
				sort.Strings(files)
				changedFiles = make(map[string]bool)
				w.callback(files)
			}

		case <-w.stop:
			return
		}
	}
}

func (w *Watcher) isRelevantFile(path string) bool {
	ext := filepath.Ext(path)
	relevantExts := []string{
		".yaml", ".yml", ".md", ".html", ".css", ".js",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot",
	}

	for _, relevant := range relevantExts {
		if ext == relevant {
			return true
		}
	}

	if strings.Contains(path, "site.yaml") {
		return true
	}

	dirs := []string{"pages", "components", "layouts", "page-types", "styles", "assets"}
	for _, dir := range dirs {
		if strings.Contains(path, dir) {
			return true
		}
	}

	return false
}

func (w *Watcher) Stop() {
	close(w.stop)
	if w.watcher != nil {
		w.watcher.Close()
	}
}
