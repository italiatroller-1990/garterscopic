package server

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher  *fsnotify.Watcher
	dir      string
	callback func()
	stop     chan struct{}
}

func NewWatcher(dir string) *Watcher {
	return &Watcher{
		dir:  dir,
		stop: make(chan struct{}),
	}
}

func (w *Watcher) Watch(callback func()) {
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
			ext := filepath.Ext(path)
			if ext == ".git" || ext == ".tmp" || ext == ".dist" {
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

	changed := false

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
					changed = true
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watch error: %v", err)

		case <-debounce.C:
			if changed {
				changed = false
				w.callback()
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
