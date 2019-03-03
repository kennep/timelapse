// Fetched from: https://raw.githubusercontent.com/bakins/logrus-middleware/master/middleware.go
// Copied instead of referenced here because the original package is no longer maintained.

package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	responseData struct {
		status int
		size   int
	}

	// loggingHandler is the actual middleware that handles logging
	loggingHandler struct {
		http.ResponseWriter
		handler      http.Handler
		responseData *responseData
	}
)

func (h *loggingHandler) newResponseData() *responseData {
	return &responseData{
		status: 0,
		size:   0,
	}
}

// Logging creates a new logging handler.
func Logging() func(http.Handler) http.Handler {
	return newLoggingHandler
}

func newLoggingHandler(next http.Handler) http.Handler {
	return &loggingHandler{
		handler: next,
	}
}

// Write is a wrapper for the "real" ResponseWriter.Write
func (h *loggingHandler) Write(b []byte) (int, error) {
	if h.responseData.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		h.responseData.status = http.StatusOK
	}
	size, err := h.ResponseWriter.Write(b)
	h.responseData.size += size
	return size, err
}

// WriteHeader is a wrapper around ResponseWriter.WriteHeader
func (h *loggingHandler) WriteHeader(s int) {
	h.ResponseWriter.WriteHeader(s)
	h.responseData.status = s
}

// Header is a wrapper around ResponseWriter.Header
func (h *loggingHandler) Header() http.Header {
	return h.ResponseWriter.Header()
}

// ServeHTTP calls the "real" handler and logs using the logger
func (h *loggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	h = newLoggingHandler(h.handler).(*loggingHandler)
	h.ResponseWriter = rw
	h.responseData = h.newResponseData()
	h.handler.ServeHTTP(h, r)

	latency := time.Since(start)

	status := h.responseData.status
	if status == 0 {
		status = 200
	}

	fields := logrus.Fields{
		"status":     status,
		"method":     r.Method,
		"request":    r.RequestURI,
		"remote":     r.RemoteAddr,
		"duration":   float64(latency.Nanoseconds()) / float64(1000),
		"size":       h.responseData.size,
		"referer":    r.Referer(),
		"user-agent": r.UserAgent(),
	}

	PopulateRequestFields(fields, r)

	logrus.WithFields(fields).Info("completed handling request")
}
