package diagnostic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/netscope/pkg/netinfo"
	"github.com/netscope/pkg/types"
)

// RunDiagnostics runs all network diagnostics on the given snapshot
func RunDiagnostics(snapshot *types.NetworkSnapshot) *types.DiagnosisReport {
	report := &types.DiagnosisReport{
		Timestamp: snapshot.Timestamp,
		Findings:  []types.DiagnosisResult{},
	}

	// Run all diagnostic checks
	diagnoseDefaultGateway(snapshot, report)
	diagnoseVPNImpact(snapshot, report)
	diagnoseDNSIssues(snapshot, report)
	diagnoseVMConnectivity(snapshot, report)
	diagnoseProxyIssues(snapshot, report)
	diagnoseMetricConflicts(snapshot, report)

	// Sort findings by severity
	sortFindingsBySeverity(report.Findings)

	// Generate summary
	report.Summary = generateSummary(report.Findings)

	return report
}

func diagnoseDefaultGateway(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	// Check if default gateway is set
	if snapshot.DefaultGateway == "" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "NO_DEFAULT_GW",
			Title:      "缺少默认网关",
			Message:    "系统没有配置默认网关，可能无法访问外部网络。",
			Suggestion: "检查网络连接，确保 DHCP 已获取或手动配置默认网关。",
		})
		return
	}

	// Check if default gateway is reachable (basic check)
	if snapshot.DefaultGateway == "0.0.0.0" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "INVALID_GW",
			Title:      "默认网关配置错误",
			Message:    "默认网关地址无效。",
			Suggestion: "重新获取网络配置或手动设置正确的网关地址。",
		})
	}
}

func diagnoseVPNImpact(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	if snapshot.VPN == nil || snapshot.VPN.Status != "connected" {
		return
	}

	// Check if VPN has a global route (0.0.0.0/0)
	for _, route := range snapshot.Routes {
		if route.Interface == snapshot.VPN.Interface {
			if route.Destination == "0.0.0.0/0" {
				report.Findings = append(report.Findings, types.DiagnosisResult{
					Severity:   "warning",
					Code:       "VPN_GLOBAL_ROUTE",
					Title:      "VPN 使用全局路由模式",
					Message:    fmt.Sprintf("VPN (%s) 配置为路由所有流量，这可能导致以下问题：\n• 访问本地网络（如虚拟机）受限\n• 网络延迟增加\n• 部分本地服务无法访问", snapshot.VPN.Name),
					Suggestion: "如果不需要全局 VPN，建议切换到「分离隧道」模式，仅对特定流量使用 VPN。",
				})
				return
			}
		}
	}
}

func diagnoseDNSIssues(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	if len(snapshot.DNS.Servers) == 0 {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "NO_DNS",
			Title:      "未配置 DNS 服务器",
			Message:    "系统没有配置任何 DNS 服务器，无法解析域名。",
			Suggestion: "在网络设置中配置 DNS 服务器，如 8.8.8.8 或 1.1.1.1。",
		})
		return
	}

	// Check for suspicious DNS
	for _, dns := range snapshot.DNS.Servers {
		// Check for DNS hijacking (unusual DNS)
		if !netinfo.IsPublicDNS(dns) && !isPrivateIP(dns) {
			report.Findings = append(report.Findings, types.DiagnosisResult{
				Severity:   "warning",
				Code:       "SUSPICIOUS_DNS",
				Title:      "检测到非标准 DNS 服务器",
				Message:    fmt.Sprintf("DNS 服务器 %s 不是已知公共 DNS，可能存在 DNS 劫持风险。", dns),
				Suggestion: "如果这不是你配置的 DNS，建议更改为公共 DNS：8.8.8.8 (Google) 或 1.1.1.1 (Cloudflare)。",
			})
		}
	}
}

func diagnoseVMConnectivity(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	if len(snapshot.VMNetworks) == 0 {
		return
	}

	// Check if routes exist for VM networks
	vmRouteFound := false
	for _, route := range snapshot.Routes {
		for _, vm := range snapshot.VMNetworks {
			if strings.Contains(route.Destination, extractSubnetPrefix(vm.Subnet)) {
				vmRouteFound = true
				break
			}
		}
		if vmRouteFound {
			break
		}
	}

	if !vmRouteFound && len(snapshot.VMNetworks) > 0 {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "warning",
			Code:       "VM_NO_ROUTE",
			Title:      "虚拟机网络可能无法访问",
			Message:    "检测到虚拟机网络适配器，但没有找到对应的路由条目。\n这可能导致宿主机无法访问虚拟机。",
			Suggestion: "检查虚拟机网络设置为 NAT 或桥接模式，并确保路由表包含 VM 网段。",
		})
	}
}

func diagnoseProxyIssues(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	if !snapshot.Proxy.HasProxy {
		return
	}

	// Check for incomplete proxy configuration
	hasHTTP := snapshot.Proxy.HTTPProxy != ""
	hasHTTPS := snapshot.Proxy.HTTPSProxy != ""

	if hasHTTP && !hasHTTPS {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "info",
			Code:       "PROXY_INCOMPLETE",
			Title:      "代理配置不完整",
			Message:    "配置了 HTTP 代理但未配置 HTTPS 代理。\n部分应用可能仍使用系统代理导致不一致。",
			Suggestion: "如果使用代理，建议同时配置 HTTP_PROXY 和 HTTPS_PROXY 环境变量。",
		})
	}

	if !hasHTTP && hasHTTPS {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "info",
			Code:       "PROXY_HTTPS_ONLY",
			Title:      "仅配置了 HTTPS 代理",
			Message:    "只配置了 HTTPS 代理但没有配置 HTTP 代理。\n部分应用可能无法正确使用代理。",
			Suggestion: "建议同时配置 HTTP_PROXY 和 HTTPS_PROXY，或使用 ALL_PROXY 统一设置。",
		})
	}
}

func diagnoseMetricConflicts(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	// Check for multiple interfaces with same metric to same destination
	gatewayMetrics := make(map[string]map[int]bool)

	for _, route := range snapshot.Routes {
		if route.Gateway == "" {
			continue
		}
		if gatewayMetrics[route.Gateway] == nil {
			gatewayMetrics[route.Gateway] = make(map[int]bool)
		}
		gatewayMetrics[route.Gateway][route.Metric] = true
	}

	for gw, metrics := range gatewayMetrics {
		if len(metrics) > 1 {
			report.Findings = append(report.Findings, types.DiagnosisResult{
				Severity:   "warning",
				Code:       "METRIC_CONFLICT",
				Title:      "检测到网关路由冲突",
				Message:    fmt.Sprintf("网关 %s 有多个相同 metric 的路由，可能导致路由不稳定。", gw),
				Suggestion: "检查网络配置，清理重复的路由规则。",
			})
		}
	}
}

func sortFindingsBySeverity(findings []types.DiagnosisResult) {
	severityOrder := map[string]int{
		"critical": 0,
		"warning":  1,
		"info":     2,
	}

	sort.Slice(findings, func(i, j int) bool {
		return severityOrder[findings[i].Severity] < severityOrder[findings[j].Severity]
	})
}

func generateSummary(findings []types.DiagnosisResult) string {
	if len(findings) == 0 {
		return "✅ 网络配置正常，未发现问题。"
	}

	critical := 0
	warning := 0
	info := 0

	for _, f := range findings {
		switch f.Severity {
		case "critical":
			critical++
		case "warning":
			warning++
		case "info":
			info++
		}
	}

	parts := []string{}
	if critical > 0 {
		parts = append(parts, fmt.Sprintf("%d 个严重问题", critical))
	}
	if warning > 0 {
		parts = append(parts, fmt.Sprintf("%d 个警告", warning))
	}
	if info > 0 {
		parts = append(parts, fmt.Sprintf("%d 个提示", info))
	}

	return fmt.Sprintf("发现 %s，建议检查。", strings.Join(parts, "，"))
}

func isPrivateIP(ip string) bool {
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
	}
	for _, block := range privateBlocks {
		if strings.HasPrefix(ip, block[:strings.Index(block, "/")]) {
			return true
		}
	}
	return false
}

func extractSubnetPrefix(subnet string) string {
	// Extract first two octets for matching
	parts := strings.Split(subnet, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return subnet
}
