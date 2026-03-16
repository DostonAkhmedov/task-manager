package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// HTTPMetricsMiddleware wraps HTTP handlers to track metrics
func HTTPMetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrappedWriter, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(wrappedWriter.statusCode)

			// Increment request counter
			metrics.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				status,
			).Inc()

			// Track request duration
			metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration)

			// Track errors
			if wrappedWriter.statusCode >= 400 {
				metrics.HTTPErrorsTotal.WithLabelValues(
					r.Method,
					r.URL.Path,
					status,
				).Inc()
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
