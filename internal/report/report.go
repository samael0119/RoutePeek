package report

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/samael0119/RoutePeek/internal/connectivity"
	"github.com/samael0119/RoutePeek/internal/diagnostic"
	"github.com/samael0119/RoutePeek/internal/discovery"
	"github.com/samael0119/RoutePeek/internal/i18n"
	"github.com/samael0119/RoutePeek/internal/overview"
	"github.com/samael0119/RoutePeek/pkg/types"
)

const Version = "0.2.0"

// Build collects and redacts the full troubleshooting package.
func Build(ctx context.Context, lang string) (*types.TroubleshootingReport, error) {
	lang = i18n.NormalizeLang(lang)
	snapshot, err := discovery.GetNetworkSnapshot()
	if err != nil {
		return nil, err
	}
	return BuildFromSnapshot(ctx, snapshot, lang), nil
}

// BuildFromSnapshot builds a report from an existing snapshot.
func BuildFromSnapshot(ctx context.Context, snapshot *types.NetworkSnapshot, lang string) *types.TroubleshootingReport {
	lang = i18n.NormalizeLang(lang)
	redactedSnapshot := RedactSnapshot(snapshot)
	diagnosisReport := diagnostic.RunDiagnosticsWithLang(redactedSnapshot, lang)
	overviewReport := overview.BuildWithLang(redactedSnapshot, diagnosisReport, lang)
	connectivityReport := redactConnectivity(connectivity.Check(ctx, snapshot, nil, lang))

	return &types.TroubleshootingReport{
		Timestamp:    time.Now(),
		Version:      Version,
		Snapshot:     redactedSnapshot,
		Diagnosis:    diagnosisReport,
		Overview:     overviewReport,
		Connectivity: connectivityReport,
		Redaction:    redactionNote(lang),
		Warnings:     []string{sensitiveWarning(lang)},
	}
}

func redactConnectivity(report *types.ConnectivityReport) *types.ConnectivityReport {
	if report == nil {
		return nil
	}
	redacted := *report
	redacted.Targets = append([]types.ConnectivityTargetResult(nil), report.Targets...)
	for i := range redacted.Targets {
		redacted.Targets[i].Target.Address = redactIPLike(redacted.Targets[i].Target.Address)
		redacted.Targets[i].DNS.Endpoint = redactIPLike(redacted.Targets[i].DNS.Endpoint)
		redacted.Targets[i].TCP.Endpoint = redactEndpoint(redacted.Targets[i].TCP.Endpoint)
		redacted.Targets[i].HTTP.Endpoint = stripProxyCredentials(redacted.Targets[i].HTTP.Endpoint)
		redacted.Targets[i].Path.Endpoint = redactIPLike(redacted.Targets[i].Path.Endpoint)
	}
	return &redacted
}

func redactEndpoint(value string) string {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return redactIPLike(value)
	}
	return net.JoinHostPort(redactIPLike(host), port)
}

func JSON(report *types.TroubleshootingReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func Markdown(report *types.TroubleshootingReport, lang string) string {
	lang = i18n.NormalizeLang(lang)
	var b strings.Builder
	writeLine(&b, "# %s", reportText(lang, "report_title"))
	writeLine(&b, "")
	writeLine(&b, "- %s: %s", reportText(lang, "report_version"), report.Version)
	writeLine(&b, "- %s: %s", reportText(lang, "report_generated"), report.Timestamp.Format(time.RFC3339))
	writeLine(&b, "- %s: %s", reportText(lang, "report_redaction"), report.Redaction)
	for _, warning := range report.Warnings {
		writeLine(&b, "- %s", warning)
	}

	writeLine(&b, "")
	writeLine(&b, "## %s", reportText(lang, "report_health"))
	if report.Overview != nil {
		writeLine(&b, "- %s: %s", reportText(lang, "report_status"), report.Overview.Health.Label)
		writeLine(&b, "- %s: %s", reportText(lang, "report_summary"), report.Overview.Health.Summary)
		if report.Overview.Health.PrimaryIssue != "" {
			writeLine(&b, "- %s: %s", reportText(lang, "report_primary_issue"), report.Overview.Health.PrimaryIssue)
		}
	}
	if report.Overview != nil && len(report.Overview.Actions) > 0 {
		action := report.Overview.Actions[0]
		writeLine(&b, "- %s: %s", reportText(lang, "report_first_advice"), action.Title)
		writeLine(&b, "  %s", action.Impact)
	}

	writeLine(&b, "")
	writeLine(&b, "## %s", reportText(lang, "report_network_state"))
	if report.Snapshot != nil {
		writeLine(&b, "- %s: %s", reportText(lang, "report_gateway"), empty(report.Snapshot.DefaultGateway, reportText(lang, "report_empty")))
		writeLine(&b, "- DNS: %s", empty(strings.Join(report.Snapshot.DNS.Servers, ", "), reportText(lang, "report_empty")))
		writeLine(&b, "- %s: %s", reportText(lang, "report_public_ip"), empty(report.Snapshot.PublicIP, reportText(lang, "report_empty")))
		writeLine(&b, "- %s: %s", reportText(lang, "report_proxy"), proxyState(report.Snapshot.Proxy, lang))
		writeLine(&b, "- VPN: %s", vpnState(report.Snapshot.VPN, lang))
	}

	writeLine(&b, "")
	writeLine(&b, "## %s", reportText(lang, "report_connectivity"))
	if report.Connectivity != nil {
		writeLine(&b, "%s", report.Connectivity.Summary)
		writeLine(&b, "")
		writeLine(&b, "| %s | DNS | TCP | HTTP | %s | %s |", reportText(lang, "report_target"), reportText(lang, "report_path"), reportText(lang, "report_conclusion"))
		writeLine(&b, "| --- | --- | --- | --- | --- | --- |")
		for _, row := range report.Connectivity.Targets {
			writeLine(&b, "| %s | %s | %s | %s | %s | %s |",
				escapePipe(row.Target.Name),
				probeCell(row.DNS),
				probeCell(row.TCP),
				probeCell(row.HTTP),
				probeCell(row.Path),
				escapePipe(row.Conclusion),
			)
		}
	}

	writeLine(&b, "")
	writeLine(&b, "## %s", reportText(lang, "report_findings"))
	if report.Diagnosis == nil || len(report.Diagnosis.Findings) == 0 {
		writeLine(&b, "- %s", reportText(lang, "report_no_findings"))
	} else {
		for _, finding := range report.Diagnosis.Findings {
			writeLine(&b, "- [%s] %s: %s", finding.Severity, finding.Title, finding.Suggestion)
		}
	}

	writeLine(&b, "")
	writeLine(&b, "## %s", reportText(lang, "report_redaction_title"))
	writeLine(&b, "%s", report.Redaction)
	writeLine(&b, "%s", sensitiveWarning(lang))
	return b.String()
}

// RedactSnapshot removes credentials and masks local identifiers while retaining troubleshooting value.
func RedactSnapshot(snapshot *types.NetworkSnapshot) *types.NetworkSnapshot {
	if snapshot == nil {
		return nil
	}
	redacted := *snapshot
	redacted.Interfaces = append([]types.NetworkInterface(nil), snapshot.Interfaces...)
	for i := range redacted.Interfaces {
		redacted.Interfaces[i].MAC = "hidden"
		redacted.Interfaces[i].IP4 = redactIPLike(redacted.Interfaces[i].IP4)
		redacted.Interfaces[i].IP6 = redactIPLike(redacted.Interfaces[i].IP6)
	}
	redacted.Routes = append([]types.RouteEntry(nil), snapshot.Routes...)
	for i := range redacted.Routes {
		redacted.Routes[i].Gateway = redactIPLike(redacted.Routes[i].Gateway)
		redacted.Routes[i].Destination = redactCIDR(redacted.Routes[i].Destination)
	}
	redacted.DNS.Servers = append([]string(nil), snapshot.DNS.Servers...)
	for i := range redacted.DNS.Servers {
		redacted.DNS.Servers[i] = redactIPLike(redacted.DNS.Servers[i])
	}
	redacted.DNS.Search = append([]string(nil), snapshot.DNS.Search...)
	redacted.Proxy.HTTPProxy = stripProxyCredentials(snapshot.Proxy.HTTPProxy)
	redacted.Proxy.HTTPSProxy = stripProxyCredentials(snapshot.Proxy.HTTPSProxy)
	redacted.Proxy.NoProxy = snapshot.Proxy.NoProxy
	if snapshot.VPN != nil {
		vpn := *snapshot.VPN
		vpn.ServerIP = redactIPLike(vpn.ServerIP)
		vpn.ClientIP = redactIPLike(vpn.ClientIP)
		redacted.VPN = &vpn
	}
	redacted.VMNetworks = append([]types.VMNetwork(nil), snapshot.VMNetworks...)
	for i := range redacted.VMNetworks {
		redacted.VMNetworks[i].Subnet = redactCIDR(redacted.VMNetworks[i].Subnet)
		redacted.VMNetworks[i].Gateway = redactIPLike(redacted.VMNetworks[i].Gateway)
		redacted.VMNetworks[i].HostIP = redactIPLike(redacted.VMNetworks[i].HostIP)
	}
	redacted.DefaultGateway = redactIPLike(snapshot.DefaultGateway)
	return &redacted
}

func redactCIDR(value string) string {
	ipPart, suffix, found := strings.Cut(value, "/")
	if !found {
		return redactIPLike(value)
	}
	return redactIPLike(ipPart) + "/" + suffix
}

func redactIPLike(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return value
	}
	if isPrivate(ip) {
		if v4 := ip.To4(); v4 != nil {
			switch {
			case v4[0] == 10:
				return "10.x.x.x"
			case v4[0] == 172:
				return fmt.Sprintf("172.%d.x.x", v4[1])
			case v4[0] == 192 && v4[1] == 168:
				return "192.168.x.x"
			default:
				return "private-ip"
			}
		}
		return "private-ipv6"
	}
	return value
}

func isPrivate(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()
}

func stripProxyCredentials(value string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User == nil {
		return value
	}
	parsed.User = nil
	return parsed.String()
}

func writeLine(b *strings.Builder, format string, args ...interface{}) {
	b.WriteString(fmt.Sprintf(format, args...))
	b.WriteByte('\n')
}

func probeCell(result types.ConnectivityProbeResult) string {
	value := result.Status
	if result.Detail != "" {
		value += ": " + result.Detail
	}
	return escapePipe(value)
}

func escapePipe(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

func empty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func proxyState(proxy types.ProxyConfig, lang string) string {
	if !proxy.HasProxy {
		return reportText(lang, "report_off")
	}
	parts := []string{}
	if proxy.HTTPProxy != "" {
		parts = append(parts, "HTTP="+proxy.HTTPProxy)
	}
	if proxy.HTTPSProxy != "" {
		parts = append(parts, "HTTPS="+proxy.HTTPSProxy)
	}
	if len(parts) == 0 {
		return reportText(lang, "report_on")
	}
	return strings.Join(parts, ", ")
}

func vpnState(vpn *types.VPNInfo, lang string) string {
	if vpn == nil || vpn.Status == "" || vpn.Status == "disconnected" {
		return reportText(lang, "report_off")
	}
	if vpn.Name != "" {
		return vpn.Status + " / " + vpn.Name
	}
	return vpn.Status
}

func redactionNote(lang string) string {
	return reportText(lang, "report_redaction_note")
}

func sensitiveWarning(lang string) string {
	return reportText(lang, "report_sensitive_warning")
}

func reportText(lang string, key string, args ...interface{}) string {
	return i18n.TFor(lang, key, args...)
}
