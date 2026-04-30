package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestHandleOverviewReturnsAggregatedReport(t *testing.T) {
	orig := getNetworkSnapshot
	t.Cleanup(func() { getNetworkSnapshot = orig })
	getNetworkSnapshot = func() (*types.NetworkSnapshot, error) {
		return &types.NetworkSnapshot{
			Timestamp:      time.Unix(1700000000, 0),
			DefaultGateway: "192.168.1.1",
			DNS:            types.DNSConfig{Servers: []string{"1.1.1.1"}},
			PublicIP:       "203.0.113.10",
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/overview", nil)
	rr := httptest.NewRecorder()

	handleOverview(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	var payload types.OverviewResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected overview JSON, got error %v and body %q", err, rr.Body.String())
	}
	if payload.Snapshot == nil || payload.Diagnosis == nil {
		t.Fatalf("expected snapshot and diagnosis, got %#v", payload)
	}
	if payload.Health.Status == "" || payload.Health.Summary == "" {
		t.Fatalf("expected health summary, got %#v", payload.Health)
	}
	if len(payload.Topology.Nodes) == 0 {
		t.Fatalf("expected topology nodes, got %#v", payload.Topology)
	}
}

func TestHandleOverviewReturnsJSONErrorWhenSnapshotFails(t *testing.T) {
	orig := getNetworkSnapshot
	t.Cleanup(func() { getNetworkSnapshot = orig })
	getNetworkSnapshot = func() (*types.NetworkSnapshot, error) {
		return nil, errors.New("snapshot unavailable")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/overview", nil)
	rr := httptest.NewRecorder()

	handleOverview(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	assertJSONError(t, rr, "overview_failed")
}

func TestHandleOverviewUsesRequestedLanguage(t *testing.T) {
	orig := getNetworkSnapshot
	t.Cleanup(func() { getNetworkSnapshot = orig })
	getNetworkSnapshot = func() (*types.NetworkSnapshot, error) {
		return &types.NetworkSnapshot{
			Timestamp:      time.Unix(1700000000, 0),
			Interfaces:     []types.NetworkInterface{{Name: "eth0", IP4: "192.168.1.20", IsUp: true}},
			Routes:         []types.RouteEntry{{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Interface: "eth0"}},
			DefaultGateway: "192.168.1.1",
			DNS:            types.DNSConfig{},
			PublicIP:       "203.0.113.10",
		}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/overview?lang=en", nil)
	rr := httptest.NewRecorder()

	handleOverview(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rr.Code, rr.Body.String())
	}
	var payload types.OverviewResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected overview JSON, got error %v and body %q", err, rr.Body.String())
	}
	if payload.Health.Label != "Likely offline" {
		t.Fatalf("expected English health copy, got %#v", payload.Health)
	}
	if payload.Diagnosis.Findings[0].Title != "No DNS Servers Configured" {
		t.Fatalf("expected English diagnostic title, got %#v", payload.Diagnosis.Findings[0])
	}
}

func TestHandleConnectivityReturnsMatrix(t *testing.T) {
	orig := getNetworkSnapshot
	origCheck := checkConnectivity
	t.Cleanup(func() {
		getNetworkSnapshot = orig
		checkConnectivity = origCheck
	})
	getNetworkSnapshot = func() (*types.NetworkSnapshot, error) {
		return &types.NetworkSnapshot{
			Timestamp:      time.Unix(1700000000, 0),
			DefaultGateway: "192.168.1.1",
			DNS:            types.DNSConfig{Servers: []string{"1.1.1.1"}},
		}, nil
	}
	checkConnectivity = func(context.Context, *types.NetworkSnapshot, []string, string) *types.ConnectivityReport {
		return &types.ConnectivityReport{
			Timestamp: time.Unix(1700000001, 0),
			Summary:   "ok",
			Targets: []types.ConnectivityTargetResult{{
				Target:     types.ConnectivityTarget{Name: "google.com", Address: "google.com"},
				Conclusion: "normal",
			}},
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/connectivity?lang=en", nil)
	rr := httptest.NewRecorder()

	handleConnectivity(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rr.Code, rr.Body.String())
	}
	var payload types.ConnectivityReport
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected connectivity JSON, got error %v and body %q", err, rr.Body.String())
	}
	if len(payload.Targets) == 0 || payload.Summary == "" {
		t.Fatalf("expected connectivity targets and summary, got %#v", payload)
	}
}

func TestHandleReportReturnsMarkdown(t *testing.T) {
	orig := getNetworkSnapshot
	origBuildReport := buildReport
	t.Cleanup(func() {
		getNetworkSnapshot = orig
		buildReport = origBuildReport
	})
	getNetworkSnapshot = func() (*types.NetworkSnapshot, error) {
		return &types.NetworkSnapshot{
			Timestamp:      time.Unix(1700000000, 0),
			DefaultGateway: "192.168.1.1",
			DNS:            types.DNSConfig{Servers: []string{"1.1.1.1"}},
			PublicIP:       "203.0.113.10",
		}, nil
	}
	buildReport = func(context.Context, *types.NetworkSnapshot, string) *types.TroubleshootingReport {
		return &types.TroubleshootingReport{
			Timestamp:    time.Unix(1700000001, 0),
			Version:      "v0.2.1",
			Redaction:    "redacted",
			Warnings:     []string{"review"},
			Overview:     &types.OverviewResponse{Health: types.HealthOverview{Label: "Healthy", Summary: "No issue"}},
			Snapshot:     &types.NetworkSnapshot{},
			Connectivity: &types.ConnectivityReport{Summary: "ok"},
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/report?format=markdown&lang=en", nil)
	rr := httptest.NewRecorder()

	handleReport(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusOK, rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "text/markdown; charset=utf-8" {
		t.Fatalf("expected markdown content type, got %q", got)
	}
	if body := rr.Body.String(); !strings.Contains(body, "# RoutePeek Troubleshooting Package") {
		t.Fatalf("unexpected markdown body %q", body)
	}
}
