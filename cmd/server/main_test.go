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

func TestHandleIPLocationReturnsLocation(t *testing.T) {
	orig := lookupIPLocation
	t.Cleanup(func() { lookupIPLocation = orig })
	lookupIPLocation = func(ip string) (*types.IPLocation, error) {
		if ip != "8.8.8.8" {
			t.Fatalf("unexpected lookup IP %q", ip)
		}
		return &types.IPLocation{Country: "United States", City: "Mountain View", Org: "Google LLC"}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/ip-location?ip=8.8.8.8", nil)
	rr := httptest.NewRecorder()

	handleIPLocation(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rr.Code, rr.Body.String())
	}
	var payload types.IPLocation
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected location JSON, got error %v and body %q", err, rr.Body.String())
	}
	if payload.Country != "United States" || payload.City != "Mountain View" || payload.Org != "Google LLC" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

func TestHandleIPLocationRejectsInvalidIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/ip-location?ip=bad;target", nil)
	rr := httptest.NewRecorder()

	handleIPLocation(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertJSONError(t, rr, "invalid_ip")
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
