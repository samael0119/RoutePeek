package netinfo

import (
	"net"
	"strings"

	"github.com/samael0119/RoutePeek/pkg/types"
)

// DetectVPN detects active VPN connections
func DetectVPN() *types.VPNInfo {
	vpn := &types.VPNInfo{
		Status: "unknown",
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return vpn
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		if looksLikeVPNInterface(iface.Name) {
			vpn.Interface = iface.Name
			vpn.Status = "connected"

			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if v, ok := addr.(*net.IPNet); ok && v.IP.To4() != nil {
					vpn.ClientIP = v.IP.String()
					vpn.Protocol = guessVPNProtocol(v.IP)
				}
			}

			return vpn
		}
	}

	routes, _ := GetRoutes()
	defaultGw, _ := GetDefaultGateway()
	networkInterfaces, _ := GetNetworkInterfaces()
	if routeVPN := vpnFromRoutes(routes, defaultGw, networkInterfaces); routeVPN != nil {
		return routeVPN
	}

	// Fallback: check routes for non-default-gateway 0.0.0.0/0 routes
	for _, route := range routes {
		if route.Destination == "default" || route.Destination == "0.0.0.0/0" {
			if route.Interface != "" && route.Gateway != defaultGw && looksLikeVPNInterface(route.Interface) {
				// This interface routes all traffic but isn't the default gateway - likely VPN
				vpn.Interface = route.Interface
				vpn.Status = "connected"
				vpn.Name = route.Interface + " (detected)"
				// Try to get client IP
				if iface, err := net.InterfaceByName(route.Interface); err == nil {
					addrs, _ := iface.Addrs()
					for _, addr := range addrs {
						if v, ok := addr.(*net.IPNet); ok && v.IP.To4() != nil {
							vpn.ClientIP = v.IP.String()
							vpn.Protocol = guessVPNProtocol(v.IP)
						}
					}
				}
				return vpn
			}
		}
	}

	vpn.Status = "disconnected"
	return vpn
}

func looksLikeVPNInterface(name string) bool {
	lower := strings.ToLower(name)
	for _, prefix := range VPNInterfacePrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	for _, keyword := range WindowsVPNInterfaceKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func vpnFromRoutes(routes []types.RouteEntry, defaultGw string, interfaces []types.NetworkInterface) *types.VPNInfo {
	for _, route := range routes {
		if route.Destination != "default" && route.Destination != "0.0.0.0/0" {
			continue
		}
		if route.Interface == "" || route.Gateway == "" || route.Gateway == defaultGw {
			continue
		}

		for _, iface := range interfaces {
			if !iface.IsUp {
				continue
			}
			if !routeMatchesInterface(route.Interface, iface) {
				continue
			}
			return &types.VPNInfo{
				Name:      iface.Name,
				Interface: iface.Name,
				Status:    "connected",
				ClientIP:  iface.IP4,
				Protocol:  guessVPNProtocol(net.ParseIP(iface.IP4)),
			}
		}
	}
	return nil
}

func routeMatchesInterface(routeInterface string, iface types.NetworkInterface) bool {
	if routeInterface == "" {
		return false
	}
	if !interfaceLooksLikeVPN(iface) && !looksLikeVPNInterface(routeInterface) {
		return false
	}
	return iface.IP4 != "" && routeInterface == iface.IP4 || strings.EqualFold(routeInterface, iface.Name)
}

func interfaceLooksLikeVPN(iface types.NetworkInterface) bool {
	return iface.Type == "vpn" || looksLikeVPNInterface(iface.Name)
}

// guessVPNProtocol identifies VPN type by IP range
func guessVPNProtocol(ip net.IP) string {
	vpnRanges := map[string]string{
		"10.0.0.0/8":     "openvpn/wireguard",
		"172.16.0.0/12":  "ipsec/vpn",
		"192.168.0.0/16": "various",
	}

	for rangeStr, proto := range vpnRanges {
		_, cidr, _ := net.ParseCIDR(rangeStr)
		if cidr != nil && cidr.Contains(ip) {
			return proto
		}
	}

	return "unknown"
}

// DetectVMNetworks detects VM network configurations
func DetectVMNetworks() []types.VMNetwork {
	var vmNetworks []types.VMNetwork

	knownSubnets := []struct {
		name   string
		subnet string
		gw     string
	}{
		{"VMware VMnet8", "192.168.56.0/24", "192.168.56.1"},
		{"VMware VMnet1", "192.168.127.0/24", "192.168.127.1"},
		{"Docker bridge", "172.17.0.0/16", "172.17.0.1"},
	}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if v, ok := addr.(*net.IPNet); ok {
				if v.IP.To4() == nil {
					continue
				}

				for _, known := range knownSubnets {
					_, subnet, _ := net.ParseCIDR(known.subnet)
					if subnet != nil && subnet.Contains(v.IP) {
						vmNetworks = append(vmNetworks, types.VMNetwork{
							Name:    known.name,
							Type:    "nat",
							Subnet:  known.subnet,
							Gateway: known.gw,
							DHCP:    true,
							HostIP:  v.IP.String(),
						})
					}
				}
			}
		}
	}

	return vmNetworks
}
