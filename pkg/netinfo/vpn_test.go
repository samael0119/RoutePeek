package netinfo

import (
	"testing"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestLooksLikeVPNInterfaceMatchesWindowsVPNNames(t *testing.T) {
	names := []string{
		"WireGuard Tunnel",
		"Tailscale",
		"Cisco AnyConnect Secure Mobility Client Virtual Miniport Adapter",
		"Fortinet SSL VPN Virtual Ethernet Adapter",
		"TAP-Windows Adapter V9",
		"OpenVPN TAP-Windows6",
		"ZeroTier One Virtual Port",
		"Cloudflare WARP",
		"PANGP Virtual Ethernet Adapter",
	}
	for _, name := range names {
		if !looksLikeVPNInterface(name) {
			t.Fatalf("expected %q to look like vpn", name)
		}
	}
}

func TestLooksLikeVPNInterfaceDoesNotMatchCommonPhysicalAdapters(t *testing.T) {
	names := []string{
		"Intel(R) Wi-Fi 6 AX201",
		"Realtek PCIe GbE Family Controller",
		"Bluetooth Device (Personal Area Network)",
		"Hyper-V Virtual Ethernet Adapter",
		"VMware Virtual Ethernet Adapter for VMnet8",
	}
	for _, name := range names {
		if looksLikeVPNInterface(name) {
			t.Fatalf("expected %q to not look like vpn", name)
		}
	}
}

func TestVPNRouteFallbackMatchesRouteInterfaceIP(t *testing.T) {
	iface := types.NetworkInterface{Name: "WireGuard Tunnel", IP4: "10.8.0.2", IsUp: true, Type: "vpn"}
	routes := []types.RouteEntry{
		{Destination: "0.0.0.0/0", Gateway: "10.8.0.1", Interface: "10.8.0.2"},
	}

	got := vpnFromRoutes(routes, "192.168.1.1", []types.NetworkInterface{iface})
	if got == nil || got.Status != "connected" || got.Interface != "WireGuard Tunnel" {
		t.Fatalf("unexpected vpn from route fallback %#v", got)
	}
}

func TestVPNRouteFallbackIgnoresDefaultGatewayRoute(t *testing.T) {
	iface := types.NetworkInterface{Name: "WireGuard Tunnel", IP4: "10.8.0.2", IsUp: true, Type: "vpn"}
	routes := []types.RouteEntry{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Interface: "10.8.0.2"},
	}

	if got := vpnFromRoutes(routes, "192.168.1.1", []types.NetworkInterface{iface}); got != nil {
		t.Fatalf("expected default gateway route to be ignored, got %#v", got)
	}
}

func TestVPNRouteFallbackIgnoresPhysicalAdapterIP(t *testing.T) {
	iface := types.NetworkInterface{Name: "Intel(R) Wi-Fi 6 AX201", IP4: "192.168.1.20", IsUp: true, Type: "wifi"}
	routes := []types.RouteEntry{
		{Destination: "0.0.0.0/0", Gateway: "192.168.99.1", Interface: "192.168.1.20"},
	}

	if got := vpnFromRoutes(routes, "192.168.1.1", []types.NetworkInterface{iface}); got != nil {
		t.Fatalf("expected physical adapter route to be ignored, got %#v", got)
	}
}
