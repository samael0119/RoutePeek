package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/routepeek/internal/diagnostic"
	"github.com/routepeek/internal/i18n"
	"github.com/routepeek/pkg/types"
)

// PrintJSON prints the given data as formatted JSON
func PrintJSON(data interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

// PrintNetworkOverview prints a comprehensive network overview
func PrintNetworkOverview(s *types.NetworkSnapshot, color bool) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  %s\n", i18n.T("title_overview"))
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Network Interfaces
	fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %s\n", i18n.T("title_iface"))
	fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
	printInterfacesTable(s.Interfaces, color)
	fmt.Println()

	// Routing
	fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %s\n", i18n.T("title_routes"))
	fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
	printRoutesTable(s.Routes, color)
	fmt.Println()

	// DNS
	fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %s\n", i18n.T("title_dns"))
	fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
	if len(s.DNS.Servers) > 0 {
		for i, dns := range s.DNS.Servers {
			suffix := ""
			if isWellKnownDNS(dns) {
				suffix = i18n.T("public_dns")
			}
			fmt.Printf("   %d. %s%s\n", i+1, dns, suffix)
		}
	} else {
		fmt.Printf("   %s\n", i18n.T("none"))
	}
	fmt.Println()

	// VPN
	if s.VPN != nil {
		fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ %s\n", i18n.T("title_vpn"))
		fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
		fmt.Printf("   %s: %s\n", i18n.T("status"), s.VPN.Status)
		if s.VPN.Status == "connected" {
			fmt.Printf("   %s: %s\n", i18n.T("interface"), s.VPN.Interface)
			fmt.Printf("   %s: %s\n", i18n.T("client_ip"), s.VPN.ClientIP)
			fmt.Printf("   %s: %s\n", i18n.T("protocol"), s.VPN.Protocol)
		}
		fmt.Println()
	}

	// Proxy
	if s.Proxy.HasProxy {
		fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ %s\n", i18n.T("title_proxy"))
		fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
		if s.Proxy.HTTPProxy != "" {
			fmt.Printf("   HTTP:  %s\n", s.Proxy.HTTPProxy)
		}
		if s.Proxy.HTTPSProxy != "" {
			fmt.Printf("   HTTPS: %s\n", s.Proxy.HTTPSProxy)
		}
		fmt.Println()
	}

	// VM Networks
	if len(s.VMNetworks) > 0 {
		fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ %s\n", i18n.T("title_vm"))
		fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
		for _, vm := range s.VMNetworks {
			fmt.Printf("   %s\n", vm.Name)
			fmt.Printf("      %s: %s | %s: %s\n", i18n.T("subnet"), vm.Subnet, i18n.T("gateway"), vm.Gateway)
		}
		fmt.Println()
	}

	// Public IP
	if s.PublicIP != "" {
		fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ %s\n", i18n.T("title_public_ip"))
		fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
		fmt.Printf("   %s\n", s.PublicIP)
		fmt.Println()
	}
}

func printInterfacesTable(ifaces []types.NetworkInterface, color bool) {
	if len(ifaces) == 0 {
		fmt.Printf("   %s\n", i18n.T("none"))
		return
	}

	for _, iface := range ifaces {
		typeIcon := getInterfaceIcon(iface.Type)
		statusIcon := "✅"
		if !iface.IsUp {
			statusIcon = "❌"
		}

		fmt.Printf("   %s %s %s\n", typeIcon, iface.Name, statusIcon)
		if iface.IP4 != "" {
			fmt.Printf("      IP:   %s\n", iface.IP4)
		}
		if iface.MAC != "" {
			fmt.Printf("      MAC:  %s\n", iface.MAC)
		}
		if iface.MTU > 0 {
			fmt.Printf("      MTU:  %d\n", iface.MTU)
		}
	}
}

func printRoutesTable(routes []types.RouteEntry, color bool) {
	if len(routes) == 0 {
		fmt.Printf("   %s\n", i18n.T("none"))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "   %-20s %-15s %-10s %s\n", i18n.T("dest_network"), i18n.T("gateway"), i18n.T("interface"), i18n.T("metric"))
	fmt.Fprintf(w, "   %-20s %-15s %-10s %s\n", "────────────────────", "───────────────", "──────────", "──────")

	for _, route := range routes {
		gw := route.Gateway
		if gw == "" {
			gw = "*"
		}
		fmt.Fprintf(w, "   %-20s %-15s %-10s %d\n", route.Destination, gw, route.Interface, route.Metric)
	}
	w.Flush()
}

func getInterfaceIcon(ifaceType string) string {
	switch ifaceType {
	case "ethernet":
		return "🔌"
	case "wifi":
		return "📶"
	case "vpn":
		return "🔒"
	case "virtual":
		return "🖥️"
	case "docker":
		return "🐳"
	case "loopback":
		return "🔁"
	default:
		return "🌐"
	}
}

func isWellKnownDNS(ip string) bool {
	knownDNS := map[string]string{
		"8.8.8.8":          "Google",
		"8.8.4.4":          "Google",
		"1.1.1.1":          "Cloudflare",
		"1.0.0.1":          "Cloudflare",
		"9.9.9.9":          "Quad9",
		"208.67.222.222":   "OpenDNS",
		"208.67.220.220":   "OpenDNS",
		"114.114.114.114":  "DNS盾",
		"114.114.115.115":  "DNS盾",
		"223.5.5.5":        "阿里",
		"223.6.6.6":        "阿里",
	}
	return knownDNS[ip] != ""
}

// PrintInterfaces prints network interfaces in a formatted way
func PrintInterfaces(ifaces []types.NetworkInterface, color bool) {
	printInterfacesTable(ifaces, color)
}

// PrintRoutes prints routing table in a formatted way
func PrintRoutes(routes []types.RouteEntry, color bool) {
	printRoutesTable(routes, color)
}

// PrintDiagnosticReport prints a formatted diagnostic report (calls diagnostic package)
func PrintDiagnosticReport(report *types.DiagnosisReport, color bool) {
	diagnostic.PrintDiagnosticReport(report, color)
}
