package connectivity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func withProbeStubs(t *testing.T) {
	t.Helper()
	origLookup := lookupHost
	origTCP := probeTCP
	origHTTP := probeHTTP
	origPath := probePath
	t.Cleanup(func() {
		lookupHost = origLookup
		probeTCP = origTCP
		probeHTTP = origHTTP
		probePath = origPath
	})
	probeTCP = func(context.Context, types.ConnectivityTarget) (string, time.Duration, error) {
		return "target:443", time.Millisecond, nil
	}
	probeHTTP = func(context.Context, types.ConnectivityTarget) (string, time.Duration, error) {
		return "https://target", time.Millisecond, nil
	}
	probePath = func(context.Context, *types.NetworkSnapshot, types.ConnectivityTarget, string) types.ConnectivityProbeResult {
		return types.ConnectivityProbeResult{Status: "info", Code: "route_probe", Detail: "route probe"}
	}
}

func TestCheckConcludesDNSFailure(t *testing.T) {
	withProbeStubs(t)
	lookupHost = func(context.Context, string) ([]string, error) {
		return nil, errors.New("no such host")
	}

	report := Check(context.Background(), &types.NetworkSnapshot{DefaultGateway: "192.168.1.1"}, []string{"example.invalid"}, "en")

	if len(report.Targets) != 1 {
		t.Fatalf("expected one target, got %#v", report.Targets)
	}
	result := report.Targets[0]
	if result.DNS.Code != "dns_failed" || result.Severity != "danger" {
		t.Fatalf("expected DNS danger, got %#v", result)
	}
	if result.Conclusion != "DNS failed. Check DNS servers, proxy DNS, or captive portal state." {
		t.Fatalf("unexpected conclusion %q", result.Conclusion)
	}
}

func TestCheckConcludesTCPFailure(t *testing.T) {
	withProbeStubs(t)
	lookupHost = func(context.Context, string) ([]string, error) {
		return []string{"93.184.216.34"}, nil
	}
	probeTCP = func(context.Context, types.ConnectivityTarget) (string, time.Duration, error) {
		return "example.com:443", 0, errors.New("refused")
	}

	report := Check(context.Background(), &types.NetworkSnapshot{DefaultGateway: "192.168.1.1"}, []string{"example.com"}, "en")

	if got := report.Targets[0].Conclusion; got != "TCP failed. Routing, firewall, proxy, or the target service may be blocking the connection." {
		t.Fatalf("unexpected conclusion %q", got)
	}
}

func TestCheckConcludesHTTPFailure(t *testing.T) {
	withProbeStubs(t)
	lookupHost = func(context.Context, string) ([]string, error) {
		return []string{"93.184.216.34"}, nil
	}
	probeHTTP = func(context.Context, types.ConnectivityTarget) (string, time.Duration, error) {
		return "https://example.com", 0, errors.New("bad status")
	}

	report := Check(context.Background(), &types.NetworkSnapshot{DefaultGateway: "192.168.1.1"}, []string{"example.com"}, "en")

	if report.Targets[0].Severity != "warning" {
		t.Fatalf("expected HTTP warning, got %#v", report.Targets[0])
	}
	if got := report.Targets[0].Conclusion; got != "TCP works but HTTP failed. The website, proxy policy, TLS interception, or HTTP status may be the issue." {
		t.Fatalf("unexpected conclusion %q", got)
	}
}

func TestCheckDetectsFakeIP(t *testing.T) {
	withProbeStubs(t)
	lookupHost = func(context.Context, string) ([]string, error) {
		return []string{"198.18.0.1"}, nil
	}

	report := Check(context.Background(), &types.NetworkSnapshot{DefaultGateway: "192.168.1.1"}, []string{"example.com"}, "en")

	if report.Targets[0].DNS.Code != "fake_ip" || report.Targets[0].Severity != "warning" {
		t.Fatalf("expected fake-IP warning, got %#v", report.Targets[0])
	}
}

func TestCheckConcludesGatewayUnavailable(t *testing.T) {
	withProbeStubs(t)
	lookupHost = func(context.Context, string) ([]string, error) {
		return []string{"93.184.216.34"}, nil
	}
	probePath = defaultPathProbe

	report := Check(context.Background(), &types.NetworkSnapshot{}, []string{"example.com"}, "en")

	if report.Targets[0].Path.Code != "no_gateway" || report.Targets[0].Severity != "danger" {
		t.Fatalf("expected no gateway danger, got %#v", report.Targets[0])
	}
}
