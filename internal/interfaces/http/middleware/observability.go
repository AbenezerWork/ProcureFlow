package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type responseRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}

	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := newResponseRecorder(w)
		start := time.Now().UTC()

		next.ServeHTTP(recorder, r)

		attrs := []any{
			"request_id", requestID(r),
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", recorder.status,
			"bytes", recorder.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		}

		if tenantID, ok := TenantFromContext(r.Context()); ok {
			attrs = append(attrs, "tenant_id", tenantID.String())
		}
		if userID, ok := AuthenticatedUserID(r.Context()); ok && userID != uuid.Nil {
			attrs = append(attrs, "user_id", userID.String())
		}

		slog.Info("http_request", attrs...)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := newResponseRecorder(w)
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error(
					"http_panic",
					"request_id", requestID(r),
					"method", r.Method,
					"path", r.URL.Path,
					"panic", recovered,
					"stack", string(debug.Stack()),
				)

				if !recorder.wroteHeader {
					http.Error(recorder, "internal server error", http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}

func requestID(r *http.Request) string {
	if requestID, ok := RequestIDFromContext(r.Context()); ok {
		return requestID
	}

	return ""
}
