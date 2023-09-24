package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// L is a logger instance, utilized throughout the application to output structured logs.
var L *logrus.Logger

const logMessage = "HTTP request" // Define a constant for the log message

func JsonLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the start time to compute the duration of the request.
		start := time.Now()

		// Wrap the response writer to capture details like status and bytes written.
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		// Log the request details.
		L.WithFields(logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.String(),
			"remote_addr": r.RemoteAddr,
			"status":      ww.Status(),
			"bytes":       ww.BytesWritten(),
			"duration":    time.Since(start).Seconds() * 1000,
			"request_id":  r.Context().Value(middleware.RequestIDKey),
		}).Info(logMessage)
	})
}

func GetLogger() *logrus.Logger {
	return L
}

func init() {
	L = logrus.New()
	L.SetFormatter(&logrus.JSONFormatter{})
}
