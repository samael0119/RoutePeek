package diagnostic

import (
	"testing"
	"time"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestRunDiagnosticsReportsMetricConflictForSameDestinationMetricDifferentInterfaces(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.Routes = []types.RouteEntry{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Interface: "eth0", Metric: 100},
		{Destination: "0.0.0.0/0", Gateway: "192.168.2.1", Interface: "wlan0", Metric: 100},
	}

	report := RunDiagnostics(snapshot)

	if !hasFinding(report, "METRIC_CONFLICT") {
		t.Fatalf("expected METRIC_CONFLICT finding, got %#v", report.Findings)
	}
}

func TestRunDiagnosticsDoesNotUseStringPrefixForVMRouteMatching(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.VMNetworks = []types.VMNetwork{
		{Name: "VirtualBox NAT", Subnet: "10.0.2.0/24"},
	}
	snapshot.Routes = []types.RouteEntry{
		{Destination: "10.0.20.0/24", Interface: "eth0"},
	}

	report := RunDiagnostics(snapshot)

	if !hasFinding(report, "VM_NO_ROUTE") {
		t.Fatalf("expected VM_NO_ROUTE finding, got %#v", report.Findings)
	}
}

func TestRunDiagnosticsAllowsMissingPublicIPAsWarning(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.PublicIP = ""

	report := RunDiagnostics(snapshot)

	if !hasFinding(report, "NO_PUBLIC_IP") {
		t.Fatalf("expected NO_PUBLIC_IP finding, got %#v", report.Findings)
	}
}

func TestRunDiagnosticsReportsNoActiveInterfaces(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.Interfaces = nil

	report := RunDiagnostics(snapshot)

	if !hasFinding(report, "NO_ACTIVE_IFACE") {
		t.Fatalf("expected NO_ACTIVE_IFACE finding, got %#v", report.Findings)
	}
}

func TestRunDiagnosticsReportsNoRoutes(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.Routes = nil

	report := RunDiagnostics(snapshot)

	if !hasFinding(report, "NO_ROUTES") {
		t.Fatalf("expected NO_ROUTES finding, got %#v", report.Findings)
	}
}

func baseSnapshot() *types.NetworkSnapshot {
	return &types.NetworkSnapshot{
		Timestamp:      time.Unix(1700000000, 0),
		DefaultGateway: "192.168.1.1",
		Interfaces: []types.NetworkInterface{
			{Name: "eth0", IP4: "192.168.1.20", IsUp: true, Type: "ethernet"},
		},
		Routes: []types.RouteEntry{
			{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Interface: "eth0", Metric: 100},
		},
		DNS: types.DNSConfig{
			Servers: []string{"1.1.1.1"},
			Port:    53,
		},
		Proxy:    types.ProxyConfig{},
		PublicIP: "203.0.113.10",
	}
}

func hasFinding(report *types.DiagnosisReport, code string) bool {
	for _, f := range report.Findings {
		if f.Code == code {
			return true
		}
	}
	return false
}
