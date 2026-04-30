package i18n

import (
	"fmt"
	"os"
	"strings"
)

var currentLang = "zh"

var dict = map[string]map[string]string{
	"zh": {
		"test_key":                   "测试",
		"cli_short":                  "RoutePeek - 网络拓扑可视化和诊断工具",
		"cli_scan":                   "扫描并显示当前网络配置",
		"cli_diag":                   "运行网络诊断",
		"cli_routes":                 "显示路由表",
		"cli_iface":                  "显示所有网络接口",
		"cli_json":                   "输出为 JSON 格式",
		"cli_nocolor":                "禁用彩色输出",
		"title_overview":             "RoutePeek - 网络配置总览",
		"title_iface":                "📡 网络接口",
		"title_routes":               "📋 路由表",
		"title_dns":                  "🔤 DNS 配置",
		"title_vpn":                  "🔒 VPN 状态",
		"title_proxy":                "🌐 代理设置",
		"title_vm":                   "🖥️  虚拟机网络",
		"title_public_ip":            "🌏 公网 IP",
		"diag_title_cli":             "网络诊断报告",
		"diag_result":                "诊断结果",
		"diag_no_issues":             "未发现任何网络问题",
		"diag_suggestion":            "建议",
		"none":                       "无",
		"diag_no_active_iface_title": "未发现活动网卡",
		"diag_no_active_iface_msg":   "系统没有可用的非回环 IPv4 网卡，可能尚未连接 Wi-Fi、有线网络或热点。",
		"diag_no_active_iface_sug":   "确认网络连接已开启，并检查网卡是否被禁用。",
		"diag_no_routes_title":       "未读取到路由表",
		"diag_no_routes_msg":         "RoutePeek 没有读取到系统路由表，可能是平台命令不可用、权限不足或网络栈尚未就绪。",
		"diag_no_routes_sug":         "确认系统网络服务正常；如果网页可用，可以先把它作为采集受限处理。",
		"diag_no_gw_title":           "缺少默认网关",
		"diag_no_gw_msg":             "系统没有配置默认网关，可能无法访问外部网络。",
		"diag_no_gw_sug":             "检查网络连接，确保 DHCP 已获取或手动配置默认网关。",
		"diag_inv_gw_title":          "默认网关配置错误",
		"diag_inv_gw_msg":            "默认网关地址无效。",
		"diag_inv_gw_sug":            "重新获取网络配置或手动设置正确的网关地址。",
		"diag_vpn_global_title":      "VPN 使用全局路由模式",
		"diag_vpn_global_msg":        "VPN (%s) 配置为路由所有流量，这可能导致访问本地网络受限、延迟增加或部分服务不可用。",
		"diag_vpn_global_sug":        "如果不需要全局 VPN，建议切换到「分离隧道」模式。",
		"diag_no_dns_title":          "未配置 DNS 服务器",
		"diag_no_dns_msg":            "系统没有配置任何 DNS 服务器，无法解析域名。",
		"diag_no_dns_sug":            "在网络设置中配置 DNS 服务器，如 8.8.8.8 或 1.1.1.1。",
		"diag_susp_dns_title":        "检测到非标准 DNS 服务器",
		"diag_susp_dns_msg":          "DNS 服务器 %s 不是已知公共 DNS，可能存在 DNS 劫持风险。",
		"diag_susp_dns_sug":          "建议更改为公共 DNS：8.8.8.8 (Google) 或 1.1.1.1 (Cloudflare)。",
		"diag_vm_route_title":        "虚拟机网络可能无法访问",
		"diag_vm_route_msg":          "检测到虚拟机网络适配器，但没有找到对应的路由条目。这可能导致宿主机无法访问虚拟机。",
		"diag_vm_route_sug":          "检查虚拟机网络设置为 NAT 或桥接模式，并确保路由表包含 VM 网段。",
		"diag_proxy_inc_title":       "代理配置不完整",
		"diag_proxy_inc_msg":         "配置了 HTTP 代理但未配置 HTTPS 代理。部分应用可能表现不一致。",
		"diag_proxy_inc_sug":         "建议同时配置 HTTP_PROXY 和 HTTPS_PROXY 环境变量。",
		"diag_gw_conflict_title":     "检测到路由 Metric 冲突",
		"diag_gw_conflict_msg":       "目标网段 %s 有多个相同 metric 的接口路由，可能导致路由不稳定。",
		"diag_gw_conflict_sug":       "检查网络配置，清理重复的路由规则。",
		"diag_multi_nic_title":       "多张网卡同时活跃",
		"diag_multi_nic_msg":         "检测到 %d 张网卡同时活跃：%s。多网卡可能导致路由选择不确定。",
		"diag_multi_nic_sug":         "可以禁用不使用的网卡，避免路由冲突。",
		"diag_no_pubip_title":        "无法获取公网 IP",
		"diag_no_pubip_msg":          "无法从外部服务获取你的公网 IP 地址。可能原因：网络未连接、防火墙阻止、或处于严格内网环境。",
		"diag_no_pubip_sug":          "检查网络连接是否正常，尝试在浏览器中访问任意网站确认。",
		"diag_summary_ok":            "✅ 网络配置正常，未发现问题。",
		"diag_summary_find":          "发现 %s，建议检查。",
		"diag_summary_crit":          "%d 个严重问题",
		"diag_summary_warn":          "%d 个警告",
		"diag_summary_info":          "%d 个提示",
		"status":                     "状态",
		"interface":                  "接口",
		"client_ip":                  "客户端 IP",
		"protocol":                   "协议",
		"subnet":                     "子网",
		"gateway":                    "网关",
		"dest_network":               "目标网络",
		"metric":                     "Metric",
		"public_dns":                 " (公共 DNS)",
	},
	"en": {
		"test_key":                   "Test",
		"cli_short":                  "RoutePeek - Network topology visualization and diagnostic tool",
		"cli_scan":                   "Scan and display current network configuration",
		"cli_diag":                   "Run network diagnostics",
		"cli_routes":                 "Show routing table",
		"cli_iface":                  "Show all network interfaces",
		"cli_json":                   "Output as JSON",
		"cli_nocolor":                "Disable colored output",
		"title_overview":             "RoutePeek - Network Configuration Overview",
		"title_iface":                "📡 Network Interfaces",
		"title_routes":               "📋 Routing Table",
		"title_dns":                  "🔤 DNS Configuration",
		"title_vpn":                  "🔒 VPN Status",
		"title_proxy":                "🌐 Proxy Settings",
		"title_vm":                   "🖥️  VM Networks",
		"title_public_ip":            "🌏 Public IP",
		"diag_title_cli":             "Diagnostic Report",
		"diag_result":                "Diagnostic Result",
		"diag_no_issues":             "No issues found",
		"diag_suggestion":            "Suggestion",
		"none":                       "None",
		"diag_no_active_iface_title": "No Active Interface Found",
		"diag_no_active_iface_msg":   "No usable non-loopback IPv4 interface was found. Wi-Fi, Ethernet, or hotspot may be disconnected.",
		"diag_no_active_iface_sug":   "Confirm the network connection is enabled and the adapter is not disabled.",
		"diag_no_routes_title":       "No Route Table Data",
		"diag_no_routes_msg":         "RoutePeek could not read route table data. A platform command may be missing, permissions may be limited, or networking may not be ready.",
		"diag_no_routes_sug":         "Check that system networking is running. If websites work, treat this as a collection limitation first.",
		"diag_no_gw_title":           "Default Gateway Missing",
		"diag_no_gw_msg":             "No default gateway configured. You may be unable to access the Internet.",
		"diag_no_gw_sug":             "Check your connection or manually configure a gateway.",
		"diag_inv_gw_title":          "Invalid Default Gateway",
		"diag_inv_gw_msg":            "The configured default gateway address is invalid.",
		"diag_inv_gw_sug":            "Re-acquire network config or set a correct gateway address.",
		"diag_vpn_global_title":      "VPN in Global Route Mode",
		"diag_vpn_global_msg":        "VPN (%s) is routing all traffic. This may limit LAN access or increase latency.",
		"diag_vpn_global_sug":        "Switch to 'Split Tunneling' if global routing is not required.",
		"diag_no_dns_title":          "No DNS Servers Configured",
		"diag_no_dns_msg":            "System has no DNS servers, unable to resolve domain names.",
		"diag_no_dns_sug":            "Configure DNS servers like 8.8.8.8 or 1.1.1.1 in network settings.",
		"diag_susp_dns_title":        "Suspicious DNS Server Detected",
		"diag_susp_dns_msg":          "DNS server %s is not a known public DNS, potential hijacking risk.",
		"diag_susp_dns_sug":          "Consider changing to public DNS: 8.8.8.8 (Google) or 1.1.1.1 (Cloudflare).",
		"diag_vm_route_title":        "VM Network May Be Unreachable",
		"diag_vm_route_msg":          "VM adapter detected but no matching route found. Host may be unable to reach VMs.",
		"diag_vm_route_sug":          "Check if VM network is in NAT/Bridge mode and verify routing table.",
		"diag_proxy_inc_title":       "Incomplete Proxy Configuration",
		"diag_proxy_inc_msg":         "HTTP proxy configured without HTTPS proxy. Some apps may behave inconsistently.",
		"diag_proxy_inc_sug":         "Set both HTTP_PROXY and HTTPS_PROXY environment variables.",
		"diag_gw_conflict_title":     "Route Metric Conflict Detected",
		"diag_gw_conflict_msg":       "Destination %s has multiple interface routes with the same metric. May cause instability.",
		"diag_gw_conflict_sug":       "Check network config and clean up redundant route rules.",
		"diag_multi_nic_title":       "Multiple Active Interfaces",
		"diag_multi_nic_msg":         "Detected %d active interfaces: %s. This may lead to routing uncertainty.",
		"diag_multi_nic_sug":         "Disable unused interfaces to avoid routing conflicts.",
		"diag_no_pubip_title":        "Unable to Get Public IP",
		"diag_no_pubip_msg":          "Could not fetch public IP from external services. Network may be restricted.",
		"diag_no_pubip_sug":          "Verify network connection or try accessing a website in your browser.",
		"diag_summary_ok":            "✅ Network configuration normal. No issues found.",
		"diag_summary_find":          "Found %s, check suggested.",
		"diag_summary_crit":          "%d critical issues",
		"diag_summary_warn":          "%d warnings",
		"diag_summary_info":          "%d tips",
		"status":                     "Status",
		"interface":                  "Interface",
		"client_ip":                  "Client IP",
		"protocol":                   "Protocol",
		"subnet":                     "Subnet",
		"gateway":                    "Gateway",
		"dest_network":               "Destination",
		"metric":                     "Metric",
		"public_dns":                 " (Public DNS)",
	},
}

// Init detects language from environment
func Init() {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LANGUAGE")
	}

	lang = strings.ToLower(lang)
	if strings.Contains(lang, "en") {
		currentLang = "en"
	} else {
		currentLang = "zh" // Default
	}
}

// T returns the translated string for the given key
func T(key string, args ...interface{}) string {
	return TFor(currentLang, key, args...)
}

// TFor returns a translated string for lang without changing global language.
func TFor(lang string, key string, args ...interface{}) string {
	lang = NormalizeLang(lang)
	val, ok := dict[lang][key]
	if !ok {
		val, ok = dict["zh"][key] // Fallback to zh
		if !ok {
			return key // Fallback to key
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

func NormalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if strings.Contains(lang, "en") {
		return "en"
	}
	return "zh"
}

func init() {
	Init()
}
