package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-zoo/bone"
)

// Application represents the main application.
type Application interface {
	// Send produces a message with fallback.
	Send(topic string, key, data []byte)
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

	s.mux.PostFunc("/", s.SendMessageHandler)

	s.mux.GetFunc("/health", s.HealthHandler)
	s.mux.NotFound(NotFoundHandler())

	return s
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type produceMessage struct {
	Topic string `json:"topic"`
	Key   string `json:"key"`
	Data  string `json:"data"`
}

// SendMessageHandler handles requests to send a message.
func (s *Server) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.app.IsHealthy(); err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	msg := produceMessage{}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&msg); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if msg.Topic == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	s.app.Send(msg.Topic, []byte(msg.Key), []byte(msg.Data))

	w.WriteHeader(200)
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
