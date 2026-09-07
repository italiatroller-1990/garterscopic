package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/garterscopic/garterscopic/internal/build"
	"github.com/garterscopic/garterscopic/internal/config"
)

type Server struct {
	Config   *config.SiteConfig
	Port     int
	Builder  BuilderInterface
	httpServer *http.Server
	watcher    *Watcher
}

type BuilderInterface interface {
	Load() error
	Build() (*build.BuildResult, error)
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
	go s.watcher.Watch(s.rebuild)

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

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
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
			fmt.Fprintf(w, "data: ping\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) rebuild() {
	log.Println("Source changed, rebuilding...")

	if err := s.Builder.Load(); err != nil {
		log.Printf("Build load error: %v", err)
		return
	}

	result, err := s.Builder.Build()
	if err != nil {
		log.Printf("Build error: %v", err)
		return
	}

	log.Printf("Build complete: %d pages generated", result.PagesGenerated)
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
