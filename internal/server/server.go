package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/italiatroller-1990/garterscopic/internal/build"
	"github.com/italiatroller-1990/garterscopic/internal/config"
)

type Server struct {
	Config     *config.SiteConfig
	Port       int
	Builder    BuilderInterface
	httpServer *http.Server
	watcher    *Watcher
	buildMu    sync.Mutex
}

type BuilderInterface interface {
	Load() error
	Build() (*build.BuildResult, error)
	RebuildRoutes(routes []string) (*build.BuildResult, error)
	PagesAffectedBy(changedFile string) []string
}

func New(cfg *config.SiteConfig, port int, builder BuilderInterface) *Server {
	return &Server{
		Config:  cfg,
		Port:    port,
		Builder: builder,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("localhost:%d", s.Port)

	mux := http.NewServeMux()

	staticDir := s.Config.Build.Output
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.MkdirAll(staticDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	mux.HandleFunc("/", s.serveFile(staticDir))
	mux.HandleFunc("/.reload", s.handleReload)

	s.watcher = NewWatcher(s.Config.Build.Source)
	go s.watcher.Watch(func(changed []string) {
		s.rebuild(changed)
	})

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Development server running at http://%s", addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	if s.watcher != nil {
		s.watcher.Stop()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) serveFile(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		filePath := filepath.Join(dir, filepath.FromSlash(path))

		// Prevent directory traversal: ensure the resolved path stays within dir.
		absDir, _ := filepath.Abs(dir)
		absFile, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFile, absDir+string(os.PathSeparator)) && absFile != absDir {
			http.NotFound(w, r)
			return
		}

		info, err := os.Stat(filePath)
		if err == nil && info.IsDir() {
			// Directory-style route: serve its index.html
			filePath = filepath.Join(filePath, "index.html")
			info, err = os.Stat(filePath)
		}
		if os.IsNotExist(err) && filepath.Ext(filePath) == "" {
			// Extensionless URL: try <path>/index.html and <path>.html
			candidates := []string{
				filepath.Join(filePath, "index.html"),
				filePath + ".html",
			}
			for _, candidate := range candidates {
				if stat, statErr := os.Stat(candidate); statErr == nil && !stat.IsDir() {
					filePath = candidate
					info, err = stat, nil
					break
				}
			}
		}
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		contentType := getContentType(filePath)
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, filePath)
	}
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, "data: ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) rebuild(changedFiles []string) {
	s.buildMu.Lock()
	defer s.buildMu.Unlock()

	log.Println("Source changed, rebuilding...")

	if err := s.Builder.Load(); err != nil {
		log.Printf("Build load error: %v", err)
		return
	}

	result, err := s.rebuildChanged(changedFiles)
	if err != nil {
		log.Printf("Build error: %v", err)
		return
	}

	if len(result.RebuiltRoutes) > 0 {
		log.Printf("Partial rebuild: %d page(s) re-rendered", len(result.RebuiltRoutes))
	} else {
		log.Printf("Build complete: %d pages generated", result.PagesGenerated)
	}
}

// rebuildChanged re-renders only the pages affected by the changed files
// when the changes are content-level (page bodies, component HTML). Global
// changes (config, styles, layouts, new/removed files) fall back to a full
// build via the builder's route-set check.
func (s *Server) rebuildChanged(changedFiles []string) (*build.BuildResult, error) {
	globalExts := map[string]bool{
		".yaml": true, ".yml": true,
	}

	var routes []string
	sawGlobal := false
	for _, file := range changedFiles {
		ext := strings.ToLower(filepath.Ext(file))
		if globalExts[ext] || strings.Contains(file, "layouts") || strings.Contains(file, "page-types") {
			sawGlobal = true
			break
		}
		routes = append(routes, s.Builder.PagesAffectedBy(file)...)
	}

	if sawGlobal || len(routes) == 0 {
		return s.Builder.Build()
	}

	// Style or asset changes are cheap to refresh alongside the pages.
	return s.Builder.RebuildRoutes(routes)
}

func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}
