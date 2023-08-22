package internal

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// L is a logger instance, utilized throughout the application to output structured logs.
var L *logrus.Logger

// JsonLogger is a middleware that wraps around HTTP handlers to provide logging capabilities.
// For every incoming request, it logs key details like the method, URL, remote address,
// response status, bytes written, duration taken, and a unique request ID. The logs
// are written in a structured JSON format to ensure easy parsing and visualization.
func JsonLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the start time to compute the duration of the request.
		start := time.Now()

		// Wrap the response writer to capture details like status and bytes written.
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		// Set the log format to JSON and log the request details.
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.WithFields(logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.String(),
			"remote_addr": r.RemoteAddr,
			"status":      ww.Status(),
			"bytes":       ww.BytesWritten(),
			"duration":    time.Since(start).Seconds() * 1000,
			"request_id":  r.Context().Value(middleware.RequestIDKey),
		}).Info("HTTP request")
	})
}

// init initializes the logger instance L and sets its format to JSON
// so that all logs produced are structured accordingly.
func init() {
	L = logrus.New()
	L.SetFormatter(&logrus.JSONFormatter{})
}
