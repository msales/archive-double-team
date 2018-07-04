package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/msales/double-team/server"
	"github.com/stretchr/testify/assert"
)

func TestServer_SendMessageHandler(t *testing.T) {
	tests := []struct {
		body string
		err  error
		code int
	}{
		{"{\"topic\":\"test\",\"data\":\"test\"}", nil, http.StatusOK},
		{"{\"data\":\"test\"}", nil, http.StatusBadRequest},
		{"{\"topic\":\"\"}", nil, http.StatusBadRequest},
		{"hello", nil, http.StatusBadRequest},
		{"", errors.New(""), http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		app := testApp{
			send: func(topic string, key, data []byte) {},
			isHealthy: func() error {
				return tt.err
			},
		}
		srv := server.New(app)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
		srv.ServeHTTP(w, req)

		assert.Equal(t, tt.code, w.Code)
	}
}

func TestServer_HealthHandler(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{nil, http.StatusOK},
		{errors.New(""), http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		app := testApp{
			isHealthy: func() error {
				return tt.err
			},
		}
		srv := server.New(app)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/health", nil)
		srv.ServeHTTP(w, req)

		assert.Equal(t, tt.code, w.Code)
	}
}

func TestNotFoundHandler(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	server.NotFoundHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

type testApp struct {
	send      func(topic string, key, data []byte)
	isHealthy func() error
}

func (a testApp) Send(topic string, key, data []byte) {
	a.send(topic, key, data)
}

func (a testApp) IsHealthy() error {
	return a.isHealthy()
}
