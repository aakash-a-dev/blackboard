package server

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mockapi/mockapi/internal/config"
	"github.com/mockapi/mockapi/internal/handler"
	"github.com/mockapi/mockapi/internal/logger"
)

// Server wraps an atomically-swappable HTTP handler for hot-reload.
type Server struct {
	active      atomic.Pointer[http.Handler]
	enableCORS  bool
	globalLatMs int
}

func New(enableCORS bool, globalLatMs int) *Server {
	return &Server{enableCORS: enableCORS, globalLatMs: globalLatMs}
}

// Load builds a new handler from cfg and atomically swaps it in.
func (s *Server) Load(cfg *config.Config) {
	h := s.buildMux(cfg)
	var iface http.Handler = h
	s.active.Store(&iface)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.enableCORS {
		applyCORS(w)
	}
	// OPTIONS preflight — decision #13
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ptr := s.active.Load()
	if ptr == nil {
		http.Error(w, "server not ready", http.StatusServiceUnavailable)
		return
	}
	(*ptr).ServeHTTP(w, r)
}

func (s *Server) buildMux(cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	for _, ep := range cfg.Endpoints {
		fullPath := joinPath(cfg.Info.BasePath, ep.Path)
		ep := ep // capture

		// Go 1.22 stdlib mux uses {param} syntax, not :param.
		goPath := colonToWildcard(fullPath)
		pattern := ep.Method + " " + goPath
		mux.Handle(pattern, handler.Build(ep, s.globalLatMs))

		logger.Info("  " + padMethod(ep.Method) + " " + fullPath)
	}

	// 404 catch-all
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Route not found"}`)) //nolint:errcheck
		logger.Request(http.StatusNotFound, r.Method, r.URL.Path, 0)
	})

	return mux
}

func applyCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// colonToWildcard converts :param segments to {param} for Go 1.22 mux.
func colonToWildcard(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			parts[i] = "{" + p[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func joinPath(base, path string) string {
	base = strings.TrimRight(base, "/")
	if base == "" {
		return path
	}
	return base + path
}

func padMethod(m string) string {
	return m + strings.Repeat(" ", 7-len(m))
}
