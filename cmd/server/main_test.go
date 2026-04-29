package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestHandleTraceRejectsInvalidTargetWithJSONError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/trace?target=example.com%3Brm", nil)
	rr := httptest.NewRecorder()

	handleTrace(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertJSONError(t, rr, "invalid_target")
}

func TestHandleTraceReturnsJSONErrorWhenTraceFails(t *testing.T) {
	orig := traceRoute
	t.Cleanup(func() { traceRoute = orig })
	traceRoute = func(string) ([]types.TraceHop, error) {
		return nil, errors.New("trace tools unavailable")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/trace?target=example.com", nil)
	rr := httptest.NewRecorder()

	handleTrace(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	assertJSONError(t, rr, "trace_failed")
}

func assertJSONError(t *testing.T, rr *httptest.ResponseRecorder, code string) {
	t.Helper()
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON body, got error %v and body %q", err, rr.Body.String())
	}
	if payload.Error.Code != code {
		t.Fatalf("expected error code %q, got %#v", code, payload.Error)
	}
	if payload.Error.Message == "" {
		t.Fatalf("expected non-empty error message")
	}
}
