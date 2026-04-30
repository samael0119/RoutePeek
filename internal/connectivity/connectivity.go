package connectivity

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/samael0119/RoutePeek/internal/i18n"
	"github.com/samael0119/RoutePeek/pkg/types"
)

const defaultTimeout = 8 * time.Second

type lookupFunc func(context.Context, string) ([]string, error)
type tcpFunc func(context.Context, types.ConnectivityTarget) (string, time.Duration, error)
type httpFunc func(context.Context, types.ConnectivityTarget) (string, time.Duration, error)
type pathFunc func(context.Context, *types.NetworkSnapshot, types.ConnectivityTarget, string) types.ConnectivityProbeResult

var (
	lookupHost lookupFunc = defaultLookupHost
	probeTCP   tcpFunc    = defaultTCPProbe
	probeHTTP  httpFunc   = defaultHTTPProbe
	probePath  pathFunc   = defaultPathProbe
)

// Check runs a bounded connectivity matrix. Empty targets use the built-in defaults.
func Check(ctx context.Context, snapshot *types.NetworkSnapshot, targets []string, lang string) *types.ConnectivityReport {
	lang = i18n.NormalizeLang(lang)
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	matrixTargets := buildTargets(snapshot, targets, lang)
	results := make([]types.ConnectivityTargetResult, len(matrixTargets))
	var wg sync.WaitGroup
	for i, target := range matrixTargets {
		wg.Add(1)
		go func(i int, target types.ConnectivityTarget) {
			defer wg.Done()
			results[i] = checkTarget(ctx, snapshot, target, lang)
		}(i, target)
	}
	wg.Wait()

	applyRegionalConclusion(results, lang)

	return &types.ConnectivityReport{
		Timestamp:  start,
		DurationMS: time.Since(start).Milliseconds(),
		Summary:    reportSummary(results, lang),
		Targets:    results,
	}
}

func buildTargets(snapshot *types.NetworkSnapshot, raw []string, lang string) []types.ConnectivityTarget {
	if len(raw) > 0 {
		targets := make([]types.ConnectivityTarget, 0, len(raw))
		for _, item := range raw {
			address := strings.TrimSpace(item)
			if address == "" {
				continue
			}
			targets = append(targets, types.ConnectivityTarget{Name: address, Address: address, Kind: targetKind(address)})
		}
		return dedupeTargets(targets)
	}

	targets := []types.ConnectivityTarget{}
	if snapshot != nil && strings.TrimSpace(snapshot.DefaultGateway) != "" {
		targets = append(targets, types.ConnectivityTarget{Name: text(lang, "conn_target_gateway"), Address: snapshot.DefaultGateway, Kind: "gateway"})
	}
	if snapshot != nil && len(snapshot.DNS.Servers) > 0 && strings.TrimSpace(snapshot.DNS.Servers[0]) != "" {
		targets = append(targets, types.ConnectivityTarget{Name: text(lang, "conn_target_dns"), Address: snapshot.DNS.Servers[0], Kind: "dns"})
	}
	targets = append(targets,
		types.ConnectivityTarget{Name: "baidu.com", Address: "baidu.com", Kind: "domestic"},
		types.ConnectivityTarget{Name: "google.com", Address: "google.com", Kind: "global"},
		types.ConnectivityTarget{Name: "8.8.8.8", Address: "8.8.8.8", Kind: "global"},
	)
	return dedupeTargets(targets)
}

func dedupeTargets(targets []types.ConnectivityTarget) []types.ConnectivityTarget {
	seen := map[string]bool{}
	deduped := make([]types.ConnectivityTarget, 0, len(targets))
	for _, target := range targets {
		key := strings.ToLower(strings.TrimSpace(target.Address))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, target)
	}
	return deduped
}

func targetKind(address string) string {
	if net.ParseIP(address) != nil {
		return "ip"
	}
	return "custom"
}

func checkTarget(ctx context.Context, snapshot *types.NetworkSnapshot, target types.ConnectivityTarget, lang string) types.ConnectivityTargetResult {
	result := types.ConnectivityTargetResult{
		Target: target,
		DNS:    checkDNS(ctx, target, lang),
		Path:   probePath(ctx, snapshot, target, lang),
	}
	result.TCP = checkTCP(ctx, target, lang)
	result.HTTP = checkHTTP(ctx, target, result.DNS, lang)
	result.Severity, result.Conclusion = conclude(result, lang)
	return result
}

func checkDNS(ctx context.Context, target types.ConnectivityTarget, lang string) types.ConnectivityProbeResult {
	if ip := net.ParseIP(target.Address); ip != nil {
		status := "ok"
		code := "ip_literal"
		detail := text(lang, "conn_dns_ip_literal")
		if isFakeIP(ip) {
			status = "warning"
			code = "fake_ip"
			detail = text(lang, "conn_dns_fake_ip")
		}
		return types.ConnectivityProbeResult{Status: status, Code: code, Detail: detail, Endpoint: ip.String()}
	}

	probeCtx, cancel := context.WithTimeout(ctx, 1200*time.Millisecond)
	defer cancel()
	start := time.Now()
	ips, err := lookupHost(probeCtx, target.Address)
	if err != nil || len(ips) == 0 {
		return types.ConnectivityProbeResult{Status: "danger", Code: "dns_failed", Detail: text(lang, "conn_dns_failed")}
	}
	sort.Strings(ips)
	status := "ok"
	code := "resolved"
	detail := fmt.Sprintf(text(lang, "conn_dns_ok"), ips[0])
	if ip := net.ParseIP(ips[0]); isFakeIP(ip) {
		status = "warning"
		code = "fake_ip"
		detail = text(lang, "conn_dns_fake_ip")
	}
	return types.ConnectivityProbeResult{Status: status, Code: code, Detail: detail, Endpoint: ips[0], LatencyMS: time.Since(start).Milliseconds()}
}

func checkTCP(ctx context.Context, target types.ConnectivityTarget, lang string) types.ConnectivityProbeResult {
	probeCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	endpoint, latency, err := probeTCP(probeCtx, target)
	if err != nil {
		return types.ConnectivityProbeResult{Status: "danger", Code: "tcp_failed", Detail: text(lang, "conn_tcp_failed"), Endpoint: endpoint}
	}
	return types.ConnectivityProbeResult{Status: "ok", Code: "tcp_ok", Detail: fmt.Sprintf(text(lang, "conn_tcp_ok"), endpoint), Endpoint: endpoint, LatencyMS: latency.Milliseconds()}
}

func checkHTTP(ctx context.Context, target types.ConnectivityTarget, dns types.ConnectivityProbeResult, lang string) types.ConnectivityProbeResult {
	if dns.Code == "dns_failed" && net.ParseIP(target.Address) == nil {
		return types.ConnectivityProbeResult{Status: "danger", Code: "http_blocked_by_dns", Detail: text(lang, "conn_http_blocked_by_dns")}
	}
	if target.Kind == "gateway" || target.Kind == "dns" || target.Kind == "ip" || net.ParseIP(target.Address) != nil {
		return types.ConnectivityProbeResult{Status: "info", Code: "http_skipped", Detail: text(lang, "conn_http_skipped")}
	}
	probeCtx, cancel := context.WithTimeout(ctx, 1800*time.Millisecond)
	defer cancel()
	endpoint, latency, err := probeHTTP(probeCtx, target)
	if err != nil {
		return types.ConnectivityProbeResult{Status: "danger", Code: "http_failed", Detail: text(lang, "conn_http_failed"), Endpoint: endpoint}
	}
	return types.ConnectivityProbeResult{Status: "ok", Code: "http_ok", Detail: fmt.Sprintf(text(lang, "conn_http_ok"), endpoint), Endpoint: endpoint, LatencyMS: latency.Milliseconds()}
}

func defaultLookupHost(ctx context.Context, target string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, target)
}

func defaultTCPProbe(ctx context.Context, target types.ConnectivityTarget) (string, time.Duration, error) {
	port := "443"
	if target.Kind == "gateway" || target.Kind == "dns" || target.Address == "8.8.8.8" {
		port = "53"
	}
	endpoint := net.JoinHostPort(target.Address, port)
	start := time.Now()
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return endpoint, 0, err
	}
	_ = conn.Close()
	return endpoint, time.Since(start), nil
}

func defaultHTTPProbe(ctx context.Context, target types.ConnectivityTarget) (string, time.Duration, error) {
	client := &http.Client{
		Timeout: 1800 * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}
	for _, scheme := range []string{"https", "http"} {
		url := scheme + "://" + target.Address
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			continue
		}
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 500 {
			return url, time.Since(start), nil
		}
		return url, time.Since(start), fmt.Errorf("http status %d", resp.StatusCode)
	}
	return "https://" + target.Address, 0, fmt.Errorf("http probe failed")
}

func defaultPathProbe(ctx context.Context, snapshot *types.NetworkSnapshot, target types.ConnectivityTarget, lang string) types.ConnectivityProbeResult {
	if err := ctx.Err(); err != nil {
		return types.ConnectivityProbeResult{Status: "danger", Code: "path_timeout", Detail: text(lang, "conn_path_timeout")}
	}
	if snapshot == nil || strings.TrimSpace(snapshot.DefaultGateway) == "" {
		return types.ConnectivityProbeResult{Status: "danger", Code: "no_gateway", Detail: text(lang, "conn_path_no_gateway")}
	}
	if target.Kind == "gateway" && target.Address == snapshot.DefaultGateway {
		return types.ConnectivityProbeResult{Status: "ok", Code: "gateway_route", Detail: text(lang, "conn_path_gateway"), Endpoint: snapshot.DefaultGateway}
	}
	return types.ConnectivityProbeResult{Status: "info", Code: "route_probe", Detail: fmt.Sprintf(text(lang, "conn_path_route_probe"), snapshot.DefaultGateway), Endpoint: snapshot.DefaultGateway}
}

func conclude(result types.ConnectivityTargetResult, lang string) (string, string) {
	switch {
	case result.DNS.Code == "fake_ip" || result.Path.Code == "fake_ip":
		return "warning", text(lang, "conn_conclusion_fake_ip")
	case result.Path.Code == "no_gateway":
		return "danger", text(lang, "conn_conclusion_gateway")
	case result.DNS.Status == "danger":
		return "danger", text(lang, "conn_conclusion_dns")
	case result.TCP.Status == "danger":
		return "danger", text(lang, "conn_conclusion_tcp")
	case result.HTTP.Status == "danger":
		return "warning", text(lang, "conn_conclusion_http")
	case result.Path.Code == "route_probe":
		return "ok", text(lang, "conn_conclusion_route_probe")
	default:
		return "ok", text(lang, "conn_conclusion_ok")
	}
}

func applyRegionalConclusion(results []types.ConnectivityTargetResult, lang string) {
	domesticOK := false
	globalBad := false
	for _, result := range results {
		if result.Target.Kind == "domestic" && result.Severity == "ok" {
			domesticOK = true
		}
		if result.Target.Kind == "global" && result.Severity == "danger" {
			globalBad = true
		}
	}
	if !domesticOK || !globalBad {
		return
	}
	for i := range results {
		if results[i].Target.Kind == "global" && results[i].Severity == "danger" {
			results[i].Conclusion = text(lang, "conn_conclusion_regional")
		}
	}
}

func reportSummary(results []types.ConnectivityTargetResult, lang string) string {
	danger := 0
	warning := 0
	for _, result := range results {
		switch result.Severity {
		case "danger":
			danger++
		case "warning":
			warning++
		}
	}
	if danger > 0 {
		return fmt.Sprintf(text(lang, "conn_summary_danger"), danger)
	}
	if warning > 0 {
		return fmt.Sprintf(text(lang, "conn_summary_warning"), warning)
	}
	return text(lang, "conn_summary_ok")
}

func isFakeIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	_, fakeIPNet, err := net.ParseCIDR("198.18.0.0/15")
	return err == nil && fakeIPNet.Contains(ip)
}

func text(lang string, key string, args ...interface{}) string {
	return i18n.TFor(lang, key, args...)
}
