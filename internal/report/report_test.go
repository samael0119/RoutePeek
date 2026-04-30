package report

import (
	"strings"
	"testing"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestRedactSnapshotHidesMACPrivateIPsAndProxyCredentials(t *testing.T) {
	snapshot := &types.NetworkSnapshot{
		Interfaces: []types.NetworkInterface{{
			Name: "eth0",
			MAC:  "00:11:22:33:44:55",
			IP4:  "192.168.1.42",
		}},
		DefaultGateway: "192.168.1.1",
		PublicIP:       "203.0.113.10",
		DNS:            types.DNSConfig{Servers: []string{"10.0.0.53"}},
		Proxy: types.ProxyConfig{
			HasProxy:   true,
			HTTPProxy:  "http://user:secret@127.0.0.1:7890",
			HTTPSProxy: "https://token@example.com:8443",
		},
		VMNetworks: []types.VMNetwork{{Subnet: "172.16.9.0/24", Gateway: "172.16.9.1", HostIP: "172.16.9.2"}},
	}

	redacted := RedactSnapshot(snapshot)

	if redacted.Interfaces[0].MAC != "hidden" {
		t.Fatalf("expected hidden MAC, got %q", redacted.Interfaces[0].MAC)
	}
	if redacted.Interfaces[0].IP4 != "192.168.x.x" || redacted.DefaultGateway != "192.168.x.x" {
		t.Fatalf("expected masked private IPs, got iface=%q gw=%q", redacted.Interfaces[0].IP4, redacted.DefaultGateway)
	}
	if redacted.PublicIP != "203.0.113.10" {
		t.Fatalf("public IP should be retained, got %q", redacted.PublicIP)
	}
	if strings.Contains(redacted.Proxy.HTTPProxy, "user") || strings.Contains(redacted.Proxy.HTTPProxy, "secret") {
		t.Fatalf("proxy credentials were not stripped: %q", redacted.Proxy.HTTPProxy)
	}
	if redacted.VMNetworks[0].Subnet != "172.16.x.x/24" {
		t.Fatalf("expected masked subnet, got %q", redacted.VMNetworks[0].Subnet)
	}
}

func TestMarkdownIncludesConnectivityAndRedactionWarning(t *testing.T) {
	payload := &types.TroubleshootingReport{
		Version:   Version,
		Redaction: "MAC hidden.",
		Warnings:  []string{"Review public IP."},
		Overview: &types.OverviewResponse{Health: types.HealthOverview{
			Label:   "Healthy",
			Summary: "No issue",
		}},
		Snapshot: &types.NetworkSnapshot{DefaultGateway: "192.168.x.x", DNS: types.DNSConfig{Servers: []string{"1.1.1.1"}}},
		Connectivity: &types.ConnectivityReport{
			Summary: "ok",
			Targets: []types.ConnectivityTargetResult{{
				Target:     types.ConnectivityTarget{Name: "google.com"},
				DNS:        types.ConnectivityProbeResult{Status: "ok", Detail: "resolved"},
				TCP:        types.ConnectivityProbeResult{Status: "ok", Detail: "connected"},
				HTTP:       types.ConnectivityProbeResult{Status: "ok", Detail: "http"},
				Path:       types.ConnectivityProbeResult{Status: "info", Detail: "route probe"},
				Conclusion: "normal",
			}},
		},
	}

	markdown := Markdown(payload, "en")

	for _, want := range []string{"# RoutePeek Troubleshooting Package", "| Target | DNS | TCP | HTTP | Path | Conclusion |", "Review public IP.", "MAC hidden."} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown missing %q:\n%s", want, markdown)
		}
	}
}
