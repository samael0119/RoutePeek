package overview

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestBuildMapsFindingsToBeginnerActions(t *testing.T) {
	tests := []struct {
		code       string
		category   string
		targetNode string
		confidence string
	}{
		{code: "NO_DNS", category: "dns", targetNode: "dns", confidence: "high"},
		{code: "NO_ACTIVE_IFACE", category: "connectivity", targetNode: "device", confidence: "high"},
		{code: "NO_ROUTES", category: "route", targetNode: "gateway", confidence: "medium"},
		{code: "SUSPICIOUS_DNS", category: "dns", targetNode: "dns", confidence: "medium"},
		{code: "VPN_GLOBAL_ROUTE", category: "vpn", targetNode: "vpn", confidence: "medium"},
		{code: "MULTI_NIC_ACTIVE", category: "route", targetNode: "gateway", confidence: "low"},
		{code: "NO_PUBLIC_IP", category: "public_ip", targetNode: "internet", confidence: "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			snapshot := overviewSnapshot()
			report := &types.DiagnosisReport{
				Timestamp: snapshot.Timestamp,
				Findings: []types.DiagnosisResult{
					{
						Severity:   severityForCode(tt.code),
						Code:       tt.code,
						Title:      tt.code + " title",
						Message:    tt.code + " message",
						Suggestion: tt.code + " suggestion",
					},
				},
			}

			result := Build(snapshot, report)

			if result.Health.PrimaryIssue == "" {
				t.Fatalf("expected primary issue to be set")
			}
			if len(result.Actions) != 1 {
				t.Fatalf("expected one action, got %#v", result.Actions)
			}
			action := result.Actions[0]
			if action.Code != tt.code {
				t.Fatalf("expected action code %q, got %#v", tt.code, action)
			}
			if action.Category != tt.category {
				t.Fatalf("expected category %q, got %#v", tt.category, action)
			}
			if action.TargetNode != tt.targetNode {
				t.Fatalf("expected target node %q, got %#v", tt.targetNode, action)
			}
			if action.Confidence != tt.confidence {
				t.Fatalf("expected confidence %q, got %#v", tt.confidence, action)
			}
			if action.Impact == "" || len(action.Steps) == 0 || action.Verify == "" {
				t.Fatalf("expected beginner guidance fields to be populated, got %#v", action)
			}
		})
	}
}

func TestBuildSummarizesHealthByMostSevereFinding(t *testing.T) {
	snapshot := overviewSnapshot()
	report := &types.DiagnosisReport{
		Timestamp: snapshot.Timestamp,
		Findings: []types.DiagnosisResult{
			{Severity: "info", Code: "MULTI_NIC_ACTIVE", Title: "Multiple networks"},
			{Severity: "critical", Code: "NO_DNS", Title: "DNS missing"},
		},
	}

	result := Build(snapshot, report)

	if result.Health.Status != "offline" {
		t.Fatalf("expected offline status, got %#v", result.Health)
	}
	if result.Health.RiskLevel != "high" {
		t.Fatalf("expected high risk, got %#v", result.Health)
	}
	if result.Health.PrimaryIssue != "DNS missing" {
		t.Fatalf("expected most severe issue title, got %#v", result.Health)
	}
	if len(result.Topology.HighlightNodes) == 0 || result.Topology.HighlightNodes[0] != "dns" {
		t.Fatalf("expected DNS highlight, got %#v", result.Topology.HighlightNodes)
	}
}

func TestBuildReportsHealthyStateWhenThereAreNoFindings(t *testing.T) {
	snapshot := overviewSnapshot()
	report := &types.DiagnosisReport{Timestamp: snapshot.Timestamp}

	result := Build(snapshot, report)

	if result.Health.Status != "ok" {
		t.Fatalf("expected ok health, got %#v", result.Health)
	}
	if result.Health.RiskLevel != "low" {
		t.Fatalf("expected low risk, got %#v", result.Health)
	}
	if result.Health.PrimaryIssue != "" {
		t.Fatalf("expected no primary issue, got %#v", result.Health)
	}
	if len(result.Actions) != 0 {
		t.Fatalf("expected no actions, got %#v", result.Actions)
	}
}

func TestBuildEncodesEmptyCollectionsAsArrays(t *testing.T) {
	snapshot := overviewSnapshot()
	report := &types.DiagnosisReport{Timestamp: snapshot.Timestamp}

	result := Build(snapshot, report)
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal overview: %v", err)
	}
	body := string(payload)

	for _, field := range []string{"findings", "actions", "highlight_nodes"} {
		if strings.Contains(body, `"`+field+`":null`) {
			t.Fatalf("expected %s to encode as an empty array, got %s", field, body)
		}
	}
}

func TestBuildWithLangReturnsEnglishOverviewCopy(t *testing.T) {
	snapshot := overviewSnapshot()
	report := &types.DiagnosisReport{
		Timestamp: snapshot.Timestamp,
		Findings: []types.DiagnosisResult{
			{Severity: "critical", Code: "NO_DNS", Title: "No DNS Servers Configured"},
		},
	}

	result := BuildWithLang(snapshot, report, "en")

	if result.Health.Label != "Likely offline" {
		t.Fatalf("expected English health label, got %#v", result.Health)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected one action, got %#v", result.Actions)
	}
	if result.Actions[0].Impact == "" || result.Actions[0].Impact == "DNS 缺失会导致输入网址时无法找到对应服务器，表现为网页打不开但直连 IP 可能可用。" {
		t.Fatalf("expected English action copy, got %#v", result.Actions[0])
	}
	if result.Topology.Nodes[0].Label != "Device" {
		t.Fatalf("expected English topology labels, got %#v", result.Topology.Nodes)
	}
}

func overviewSnapshot() *types.NetworkSnapshot {
	return &types.NetworkSnapshot{
		Timestamp: time.Unix(1700000000, 0),
		Interfaces: []types.NetworkInterface{
			{Name: "eth0", IP4: "192.168.1.20", IsUp: true, Type: "ethernet"},
		},
		Routes: []types.RouteEntry{
			{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Interface: "eth0", Metric: 100},
		},
		DNS:            types.DNSConfig{Servers: []string{"1.1.1.1"}, Port: 53},
		DefaultGateway: "192.168.1.1",
		PublicIP:       "203.0.113.10",
	}
}

func severityForCode(code string) string {
	if code == "NO_DNS" {
		return "critical"
	}
	if code == "MULTI_NIC_ACTIVE" {
		return "info"
	}
	return "warning"
}
