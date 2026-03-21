package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type logBuffer struct {
	builder strings.Builder
}

func (b *logBuffer) Write(p []byte) (int, error) {
	return b.builder.Write(p)
}

func TestWithRequestIDGeneratesResponseHeader(t *testing.T) {
	t.Parallel()

	handler := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := RequestIDFromContext(r.Context())
		if !ok || requestID == "" {
			t.Fatal("expected request id in context")
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if got := recorder.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected response request id header")
	}
}

func TestWithRequestIDPreservesIncomingHeader(t *testing.T) {
	t.Parallel()

	handler := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := RequestIDFromContext(r.Context())
		if !ok || requestID != "req-123" {
			t.Fatalf("expected request id req-123, got %q", requestID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(RequestIDHeader, "req-123")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(RequestIDHeader); got != "req-123" {
		t.Fatalf("expected response request id req-123, got %q", got)
	}
}

func TestRecoverReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	handler := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestAccessLogWritesStructuredEntry(t *testing.T) {
	t.Parallel()

	var logs logBuffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	previous := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(previous)

	handler := AccessLog(WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	})))

	request := httptest.NewRequest(http.MethodGet, "/healthz?check=1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	output := logs.builder.String()
	if !strings.Contains(output, "http_request") {
		t.Fatalf("expected access log entry, got %q", output)
	}
	if !strings.Contains(output, "status=202") {
		t.Fatalf("expected status in access log, got %q", output)
	}
	if !strings.Contains(output, "path=/healthz") {
		t.Fatalf("expected path in access log, got %q", output)
	}
}
