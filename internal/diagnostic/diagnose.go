package diagnostic

import (
	"fmt"
	"net"
	"sort"
	"strings"
 
	"github.com/routepeek/internal/i18n"

	"github.com/routepeek/pkg/netinfo"
	"github.com/routepeek/pkg/types"
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
	diagnoseMultiNICConflict(snapshot, report)
	diagnoseNoPublicIP(snapshot, report)

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
			Title:      i18n.T("diag_no_gw_title"),
			Message:    i18n.T("diag_no_gw_msg"),
			Suggestion: i18n.T("diag_no_gw_sug"),
		})
		return
	}

	// Check if default gateway is reachable (basic check)
	if snapshot.DefaultGateway == "0.0.0.0" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "INVALID_GW",
			Title:      i18n.T("diag_inv_gw_title"),
			Message:    i18n.T("diag_inv_gw_msg"),
			Suggestion: i18n.T("diag_inv_gw_sug"),
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
					Title:      i18n.T("diag_vpn_global_title"),
					Message:    fmt.Sprintf(i18n.T("diag_vpn_global_msg"), snapshot.VPN.Name),
					Suggestion: i18n.T("diag_vpn_global_sug"),
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
			Title:      i18n.T("diag_no_dns_title"),
			Message:    i18n.T("diag_no_dns_msg"),
			Suggestion: i18n.T("diag_no_dns_sug"),
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
				Title:      i18n.T("diag_susp_dns_title"),
				Message:    fmt.Sprintf(i18n.T("diag_susp_dns_msg"), dns),
				Suggestion: i18n.T("diag_susp_dns_sug"),
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
			Title:      i18n.T("diag_vm_route_title"),
			Message:    i18n.T("diag_vm_route_msg"),
			Suggestion: i18n.T("diag_vm_route_sug"),
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
			Title:      i18n.T("diag_proxy_inc_title"),
			Message:    i18n.T("diag_proxy_inc_msg"),
			Suggestion: i18n.T("diag_proxy_inc_sug"),
		})
	}

	if !hasHTTP && hasHTTPS {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "info",
			Code:       "PROXY_HTTPS_ONLY",
			Title:      i18n.T("diag_proxy_inc_title"),
			Message:    i18n.T("diag_proxy_inc_msg"),
			Suggestion: i18n.T("diag_proxy_inc_sug"),
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
				Title:      i18n.T("diag_gw_conflict_title"),
				Message:    fmt.Sprintf(i18n.T("diag_gw_conflict_msg"), gw),
				Suggestion: i18n.T("diag_gw_conflict_sug"),
			})
		}
	}
}

func diagnoseMultiNICConflict(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	// Count active non-loopback, non-vm interfaces
	activeIfaces := []string{}
	for _, iface := range snapshot.Interfaces {
		if iface.IsUp && !iface.IsLoopback && iface.IP4 != "" &&
			iface.Type != "vm" && iface.Type != "docker" && iface.Type != "loopback" {
			activeIfaces = append(activeIfaces, iface.Name)
		}
	}
	if len(activeIfaces) > 1 {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "info",
			Code:       "MULTI_NIC_ACTIVE",
			Title:      i18n.T("diag_multi_nic_title"),
			Message:    fmt.Sprintf(i18n.T("diag_multi_nic_msg"), len(activeIfaces), strings.Join(activeIfaces, ", ")),
			Suggestion: i18n.T("diag_multi_nic_sug"),
		})
	}
}

func diagnoseNoPublicIP(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
	if snapshot.PublicIP == "" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "warning",
			Code:       "NO_PUBLIC_IP",
			Title:      i18n.T("diag_no_pubip_title"),
			Message:    i18n.T("diag_no_pubip_msg"),
			Suggestion: i18n.T("diag_no_pubip_sug"),
		})
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
		return i18n.T("diag_summary_ok")
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
		parts = append(parts, fmt.Sprintf(i18n.T("diag_summary_crit"), critical))
	}
	if warning > 0 {
		parts = append(parts, fmt.Sprintf(i18n.T("diag_summary_warn"), warning))
	}
	if info > 0 {
		parts = append(parts, fmt.Sprintf(i18n.T("diag_summary_info"), info))
	}

	return fmt.Sprintf(i18n.T("diag_summary_find"), strings.Join(parts, ", "))
}

func isPrivateIP(ip string) bool {
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"127.0.0.0/8",
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	for _, block := range privateBlocks {
		_, cidr, err := net.ParseCIDR(block)
		if err == nil && cidr.Contains(parsedIP) {
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
