package i18n

import (
	"fmt"
	"os"
	"strings"
)

var currentLang = "zh"

var dict = map[string]map[string]string{
	"zh": {
		"test_key":        "测试",
		"cli_short":       "RoutePeek - 网络拓扑可视化和诊断工具",
		"cli_scan":        "扫描并显示当前网络配置",
		"cli_diag":        "运行网络诊断",
		"cli_routes":      "显示路由表",
		"cli_iface":       "显示所有网络接口",
		"cli_json":        "输出为 JSON 格式",
		"cli_nocolor":     "禁用彩色输出",
		"title_overview":  "RoutePeek - 网络配置总览",
		"title_iface":     "📡 网络接口",
		"title_routes":    "📋 路由表",
		"title_dns":       "🔤 DNS 配置",
		"title_vpn":       "🔒 VPN 状态",
		"title_proxy":     "🌐 代理设置",
		"title_vm":        "🖥️  虚拟机网络",
		"title_public_ip": "🌏 公网 IP",
		"none":            "无",
		"status":          "状态",
		"interface":       "接口",
		"client_ip":       "客户端 IP",
		"protocol":        "协议",
		"subnet":          "子网",
		"gateway":         "网关",
		"dest_network":    "目标网络",
		"metric":          "Metric",
		"public_dns":      " (公共 DNS)",
	},
	"en": {
		"test_key":        "Test",
		"cli_short":       "RoutePeek - Network topology visualization and diagnostic tool",
		"cli_scan":        "Scan and display current network configuration",
		"cli_diag":        "Run network diagnostics",
		"cli_routes":      "Show routing table",
		"cli_iface":       "Show all network interfaces",
		"cli_json":        "Output as JSON",
		"cli_nocolor":     "Disable colored output",
		"title_overview":  "RoutePeek - Network Configuration Overview",
		"title_iface":     "📡 Network Interfaces",
		"title_routes":    "📋 Routing Table",
		"title_dns":       "🔤 DNS Configuration",
		"title_vpn":       "🔒 VPN Status",
		"title_proxy":     "🌐 Proxy Settings",
		"title_vm":        "🖥️  VM Networks",
		"title_public_ip": "🌏 Public IP",
		"none":            "None",
		"status":          "Status",
		"interface":       "Interface",
		"client_ip":       "Client IP",
		"protocol":        "Protocol",
		"subnet":          "Subnet",
		"gateway":         "Gateway",
		"dest_network":    "Destination",
		"metric":          "Metric",
		"public_dns":      " (Public DNS)",
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
	val, ok := dict[currentLang][key]
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

func init() {
	Init()
}
