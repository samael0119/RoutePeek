package diagnostic

import (
	"net"
	"sort"
	"strings"

	"github.com/samael0119/RoutePeek/internal/i18n"

	"github.com/samael0119/RoutePeek/pkg/netinfo"
	"github.com/samael0119/RoutePeek/pkg/types"
)

type translateFunc func(string, ...interface{}) string

// RunDiagnostics runs all network diagnostics on the given snapshot
func RunDiagnostics(snapshot *types.NetworkSnapshot) *types.DiagnosisReport {
	return runDiagnostics(snapshot, i18n.T)
}

// RunDiagnosticsWithLang runs diagnostics with an explicit language.
func RunDiagnosticsWithLang(snapshot *types.NetworkSnapshot, lang string) *types.DiagnosisReport {
	return runDiagnostics(snapshot, func(key string, args ...interface{}) string {
		return i18n.TFor(lang, key, args...)
	})
}

func runDiagnostics(snapshot *types.NetworkSnapshot, t translateFunc) *types.DiagnosisReport {
	report := &types.DiagnosisReport{
		Timestamp: snapshot.Timestamp,
		Findings:  []types.DiagnosisResult{},
	}

	// Run all diagnostic checks
	diagnoseInterfaceState(snapshot, report, t)
	diagnoseRouteState(snapshot, report, t)
	diagnoseDefaultGateway(snapshot, report, t)
	diagnoseVPNImpact(snapshot, report, t)
	diagnoseDNSIssues(snapshot, report, t)
	diagnoseVMConnectivity(snapshot, report, t)
	diagnoseProxyIssues(snapshot, report, t)
	diagnoseMetricConflicts(snapshot, report, t)
	diagnoseMultiNICConflict(snapshot, report, t)
	diagnoseNoPublicIP(snapshot, report, t)

	// Sort findings by severity
	sortFindingsBySeverity(report.Findings)

	// Generate summary
	report.Summary = generateSummary(report.Findings, t)

	return report
}

func diagnoseInterfaceState(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	for _, iface := range snapshot.Interfaces {
		if iface.IsUp && !iface.IsLoopback && iface.IP4 != "" {
			return
		}
	}
	report.Findings = append(report.Findings, types.DiagnosisResult{
		Severity:   "critical",
		Code:       "NO_ACTIVE_IFACE",
		Title:      t("diag_no_active_iface_title"),
		Message:    t("diag_no_active_iface_msg"),
		Suggestion: t("diag_no_active_iface_sug"),
	})
}

func diagnoseRouteState(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	if len(snapshot.Routes) > 0 {
		return
	}
	report.Findings = append(report.Findings, types.DiagnosisResult{
		Severity:   "warning",
		Code:       "NO_ROUTES",
		Title:      t("diag_no_routes_title"),
		Message:    t("diag_no_routes_msg"),
		Suggestion: t("diag_no_routes_sug"),
	})
}

func diagnoseDefaultGateway(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	// Check if default gateway is set
	if snapshot.DefaultGateway == "" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "NO_DEFAULT_GW",
			Title:      t("diag_no_gw_title"),
			Message:    t("diag_no_gw_msg"),
			Suggestion: t("diag_no_gw_sug"),
		})
		return
	}

	// Check if default gateway is reachable (basic check)
	if snapshot.DefaultGateway == "0.0.0.0" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "INVALID_GW",
			Title:      t("diag_inv_gw_title"),
			Message:    t("diag_inv_gw_msg"),
			Suggestion: t("diag_inv_gw_sug"),
		})
	}
}

func diagnoseVPNImpact(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
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
					Title:      t("diag_vpn_global_title"),
					Message:    t("diag_vpn_global_msg", snapshot.VPN.Name),
					Suggestion: t("diag_vpn_global_sug"),
				})
				return
			}
		}
	}
}

func diagnoseDNSIssues(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	if len(snapshot.DNS.Servers) == 0 {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "critical",
			Code:       "NO_DNS",
			Title:      t("diag_no_dns_title"),
			Message:    t("diag_no_dns_msg"),
			Suggestion: t("diag_no_dns_sug"),
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
				Title:      t("diag_susp_dns_title"),
				Message:    t("diag_susp_dns_msg", dns),
				Suggestion: t("diag_susp_dns_sug"),
			})
		}
	}
}

func diagnoseVMConnectivity(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	if len(snapshot.VMNetworks) == 0 {
		return
	}

	// Check if routes exist for VM networks
	vmRouteFound := false
	for _, route := range snapshot.Routes {
		for _, vm := range snapshot.VMNetworks {
			if routeMatchesSubnet(route.Destination, vm.Subnet) {
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
			Title:      t("diag_vm_route_title"),
			Message:    t("diag_vm_route_msg"),
			Suggestion: t("diag_vm_route_sug"),
		})
	}
}

func diagnoseProxyIssues(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
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
			Title:      t("diag_proxy_inc_title"),
			Message:    t("diag_proxy_inc_msg"),
			Suggestion: t("diag_proxy_inc_sug"),
		})
	}

	if !hasHTTP && hasHTTPS {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "info",
			Code:       "PROXY_HTTPS_ONLY",
			Title:      t("diag_proxy_inc_title"),
			Message:    t("diag_proxy_inc_msg"),
			Suggestion: t("diag_proxy_inc_sug"),
		})
	}
}

func diagnoseMetricConflicts(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	type routeMetricKey struct {
		destination string
		metric      int
	}

	routesByDestinationMetric := make(map[routeMetricKey]map[string]bool)

	for _, route := range snapshot.Routes {
		if route.Destination == "" || route.Interface == "" {
			continue
		}
		key := routeMetricKey{destination: route.Destination, metric: route.Metric}
		if routesByDestinationMetric[key] == nil {
			routesByDestinationMetric[key] = make(map[string]bool)
		}
		routesByDestinationMetric[key][route.Interface] = true
	}

	for key, interfaces := range routesByDestinationMetric {
		if len(interfaces) > 1 {
			report.Findings = append(report.Findings, types.DiagnosisResult{
				Severity:   "warning",
				Code:       "METRIC_CONFLICT",
				Title:      t("diag_gw_conflict_title"),
				Message:    t("diag_gw_conflict_msg", key.destination),
				Suggestion: t("diag_gw_conflict_sug"),
			})
		}
	}
}

func diagnoseMultiNICConflict(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
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
			Title:      t("diag_multi_nic_title"),
			Message:    t("diag_multi_nic_msg", len(activeIfaces), strings.Join(activeIfaces, ", ")),
			Suggestion: t("diag_multi_nic_sug"),
		})
	}
}

func diagnoseNoPublicIP(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, t translateFunc) {
	if snapshot.PublicIP == "" {
		report.Findings = append(report.Findings, types.DiagnosisResult{
			Severity:   "warning",
			Code:       "NO_PUBLIC_IP",
			Title:      t("diag_no_pubip_title"),
			Message:    t("diag_no_pubip_msg"),
			Suggestion: t("diag_no_pubip_sug"),
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

func generateSummary(findings []types.DiagnosisResult, t translateFunc) string {
	if len(findings) == 0 {
		return t("diag_summary_ok")
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
		parts = append(parts, t("diag_summary_crit", critical))
	}
	if warning > 0 {
		parts = append(parts, t("diag_summary_warn", warning))
	}
	if info > 0 {
		parts = append(parts, t("diag_summary_info", info))
	}

	return t("diag_summary_find", strings.Join(parts, ", "))
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

func routeMatchesSubnet(routeDestination, vmSubnet string) bool {
	_, routeNet, routeErr := net.ParseCIDR(routeDestination)
	_, vmNet, vmErr := net.ParseCIDR(vmSubnet)
	if routeErr != nil || vmErr != nil || routeNet == nil || vmNet == nil {
		return false
	}
	return networkContains(routeNet, vmNet) || networkContains(vmNet, routeNet)
}

func networkContains(parent, child *net.IPNet) bool {
	if parent == nil || child == nil {
		return false
	}
	if !parent.Contains(child.IP) {
		return false
	}
	parentOnes, parentBits := parent.Mask.Size()
	childOnes, childBits := child.Mask.Size()
	return parentBits == childBits && parentOnes <= childOnes
}
