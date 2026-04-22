package netinfo

import (
	"net"
	"strings"

	"github.com/netscope/pkg/types"
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

		for _, prefix := range VPNInterfacePrefixes {
			if strings.HasPrefix(strings.ToLower(iface.Name), prefix) {
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
	}

	vpn.Status = "disconnected"
	return vpn
}

// guessVPNProtocol identifies VPN type by IP range
func guessVPNProtocol(ip net.IP) string {
	vpnRanges := map[string]string{
		"10.0.0.0/8":    "openvpn/wireguard",
		"172.16.0.0/12": "ipsec/vpn",
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
