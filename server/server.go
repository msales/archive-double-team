package server

import (
	"net/http"

	"github.com/go-zoo/bone"
)

// Application represents the main application.
type Application interface {
	// IsHealthy checks the health of the Application.
	IsHealthy() error
}

// Server represents a http server handler.
type Server struct {
	app Application
	mux *bone.Mux
}

// New creates a new Server instance.
func New(app Application) *Server {
	s := &Server{
		app: app,
		mux: bone.New(),
	}

	//s.mux.GetFunc("/:group/:file", s.ImageHandler)

	s.mux.GetFunc("/health", s.HealthHandler)
	s.mux.NotFound(NotFoundHandler())

	return s
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// HealthHandler handles health requests.
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.app.IsHealthy(); err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// NotFoundHandler returns a 404.
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
