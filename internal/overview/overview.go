package overview

import (
	"sort"

	"github.com/samael0119/RoutePeek/internal/i18n"
	"github.com/samael0119/RoutePeek/pkg/types"
)

type guidanceTemplate struct {
	category   string
	impact     string
	steps      []string
	verify     string
	targetNode string
	confidence string
}

// Build converts low-level discovery and diagnostic data into a UI-ready model.
func Build(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) *types.OverviewResponse {
	return BuildWithLang(snapshot, report, "zh")
}

// BuildWithLang converts low-level data into a UI-ready model in an explicit language.
func BuildWithLang(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport, lang string) *types.OverviewResponse {
	lang = i18n.NormalizeLang(lang)
	if report == nil {
		report = &types.DiagnosisReport{}
	}

	findings := append([]types.DiagnosisResult(nil), report.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})

	actions := make([]types.ActionItem, 0, len(findings))
	highlights := make([]string, 0, len(findings))
	seenHighlights := map[string]bool{}
	for _, finding := range findings {
		action := actionForFinding(finding, lang)
		actions = append(actions, action)
		if action.TargetNode != "" && !seenHighlights[action.TargetNode] {
			highlights = append(highlights, action.TargetNode)
			seenHighlights[action.TargetNode] = true
		}
	}

	return &types.OverviewResponse{
		Snapshot:  snapshot,
		Diagnosis: report,
		Health:    buildHealth(findings, lang),
		Actions:   actions,
		Topology:  buildTopology(snapshot, highlights, lang),
	}
}

func buildHealth(findings []types.DiagnosisResult, lang string) types.HealthOverview {
	if len(findings) == 0 {
		return types.HealthOverview{
			Status:    "ok",
			RiskLevel: "low",
			Label:     overviewText(lang, "health_ok_label"),
			Summary:   overviewText(lang, "health_ok_summary"),
		}
	}

	primary := findings[0]
	switch primary.Severity {
	case "critical":
		return types.HealthOverview{
			Status:       "offline",
			RiskLevel:    "high",
			Label:        overviewText(lang, "health_offline_label"),
			PrimaryIssue: primary.Title,
			Summary:      overviewText(lang, "health_offline_summary"),
		}
	case "warning":
		return types.HealthOverview{
			Status:       "attention",
			RiskLevel:    "medium",
			Label:        overviewText(lang, "health_attention_label"),
			PrimaryIssue: primary.Title,
			Summary:      overviewText(lang, "health_attention_summary"),
		}
	default:
		return types.HealthOverview{
			Status:       "ok",
			RiskLevel:    "low",
			Label:        overviewText(lang, "health_ok_label"),
			PrimaryIssue: primary.Title,
			Summary:      overviewText(lang, "health_info_summary"),
		}
	}
}

func actionForFinding(finding types.DiagnosisResult, lang string) types.ActionItem {
	template := guidanceForCode(finding.Code, lang)
	return types.ActionItem{
		Code:       finding.Code,
		Severity:   finding.Severity,
		Title:      finding.Title,
		Category:   template.category,
		Impact:     template.impact,
		Steps:      append([]string(nil), template.steps...),
		Verify:     template.verify,
		TargetNode: template.targetNode,
		Confidence: template.confidence,
	}
}

func guidanceForCode(code string, lang string) guidanceTemplate {
	templates := guidanceTemplates(lang)
	if template, ok := templates[code]; ok {
		return template
	}
	return guidanceTemplate{
		category:   "connectivity",
		impact:     overviewText(lang, "action_default_impact"),
		steps:      []string{overviewText(lang, "action_default_step1"), overviewText(lang, "action_default_step2"), overviewText(lang, "action_default_step3")},
		verify:     overviewText(lang, "action_default_verify"),
		targetNode: "gateway",
		confidence: "low",
	}
}

func buildTopology(snapshot *types.NetworkSnapshot, highlights []string, lang string) types.NetworkTopology {
	highlightSet := map[string]bool{}
	for _, id := range highlights {
		highlightSet[id] = true
	}

	nodes := []types.TopologyNode{
		{ID: "device", Label: overviewText(lang, "node_device"), Kind: "device", Status: "ok", Description: activeInterfaceDescription(snapshot, lang)},
		{ID: "gateway", Label: overviewText(lang, "node_gateway"), Kind: "gateway", Status: nodeStatus("gateway", highlightSet), Description: gatewayDescription(snapshot, lang)},
		{ID: "dns", Label: "DNS", Kind: "dns", Status: nodeStatus("dns", highlightSet), Description: dnsDescription(snapshot, lang)},
		{ID: "internet", Label: overviewText(lang, "node_internet"), Kind: "internet", Status: nodeStatus("internet", highlightSet), Description: publicIPDescription(snapshot, lang)},
	}
	links := []types.TopologyLink{
		{From: "device", To: "gateway", Status: linkStatus("gateway", highlightSet), Label: overviewText(lang, "link_gateway")},
		{From: "gateway", To: "internet", Status: linkStatus("internet", highlightSet), Label: overviewText(lang, "link_internet")},
		{From: "device", To: "dns", Status: linkStatus("dns", highlightSet), Label: overviewText(lang, "link_dns")},
	}

	if snapshot != nil && snapshot.Proxy.HasProxy {
		nodes = append(nodes, types.TopologyNode{ID: "proxy", Label: overviewText(lang, "node_proxy"), Kind: "proxy", Status: nodeStatus("proxy", highlightSet), Description: overviewText(lang, "desc_proxy_on")})
		links = append(links, types.TopologyLink{From: "device", To: "proxy", Status: linkStatus("proxy", highlightSet), Label: overviewText(lang, "link_proxy")})
	}
	if snapshot != nil && snapshot.VPN != nil && snapshot.VPN.Status == "connected" {
		nodes = append(nodes, types.TopologyNode{ID: "vpn", Label: "VPN", Kind: "vpn", Status: nodeStatus("vpn", highlightSet), Description: snapshot.VPN.Name})
		links = append(links, types.TopologyLink{From: "device", To: "vpn", Status: linkStatus("vpn", highlightSet), Label: overviewText(lang, "link_vpn")})
	}
	if snapshot != nil && len(snapshot.VMNetworks) > 0 {
		nodes = append(nodes, types.TopologyNode{ID: "vm", Label: overviewText(lang, "node_vm"), Kind: "vm", Status: nodeStatus("vm", highlightSet), Description: snapshot.VMNetworks[0].Subnet})
		links = append(links, types.TopologyLink{From: "device", To: "vm", Status: linkStatus("vm", highlightSet), Label: overviewText(lang, "link_vm")})
	}

	return types.NetworkTopology{
		Nodes:          nodes,
		Links:          links,
		HighlightNodes: highlights,
	}
}

func nodeStatus(id string, highlights map[string]bool) string {
	if highlights[id] {
		return "risk"
	}
	return "ok"
}

func linkStatus(id string, highlights map[string]bool) string {
	if highlights[id] {
		return "risk"
	}
	return "ok"
}

func activeInterfaceDescription(snapshot *types.NetworkSnapshot, lang string) string {
	if snapshot == nil {
		return overviewText(lang, "desc_waiting_snapshot")
	}
	for _, iface := range snapshot.Interfaces {
		if iface.IsUp && !iface.IsLoopback && iface.IP4 != "" {
			return iface.Name + " / " + iface.IP4
		}
	}
	return overviewText(lang, "desc_no_active_iface")
}

func gatewayDescription(snapshot *types.NetworkSnapshot, lang string) string {
	if snapshot == nil || snapshot.DefaultGateway == "" {
		return overviewText(lang, "desc_no_gateway")
	}
	return snapshot.DefaultGateway
}

func dnsDescription(snapshot *types.NetworkSnapshot, lang string) string {
	if snapshot == nil || len(snapshot.DNS.Servers) == 0 {
		return overviewText(lang, "desc_no_dns")
	}
	return snapshot.DNS.Servers[0]
}

func publicIPDescription(snapshot *types.NetworkSnapshot, lang string) string {
	if snapshot == nil || snapshot.PublicIP == "" {
		return overviewText(lang, "desc_no_public_ip")
	}
	return snapshot.PublicIP
}

func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}
