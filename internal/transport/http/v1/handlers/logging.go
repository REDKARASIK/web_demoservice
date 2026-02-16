package handlers

import (
	"log/slog"
	"net/http"
	"time"
)

type LoggingOrderHandler struct {
	next OrderHTTPHandler
}

func NewLoggingOrderHandler(next OrderHTTPHandler) *LoggingOrderHandler {
	return &LoggingOrderHandler{next: next}
}

func (h *LoggingOrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

	h.next.GetOrder(rec, r)

	slog.Info(
		"http request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", rec.status),
		slog.Duration("duration", time.Since(start)),
	)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
