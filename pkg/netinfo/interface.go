package netinfo

import (
	"fmt"
	"net"
	"strings"

	"github.com/netscope/pkg/types"
)

// GetNetworkInterfaces returns all network interfaces on the system
func GetNetworkInterfaces() ([]types.NetworkInterface, error) {
	var interfaces []types.NetworkInterface

	ifaceList, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, iface := range ifaceList {
		// Skip loopback for now
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		ni := types.NetworkInterface{
			Name:       iface.Name,
			Index:      iface.Index,
			MTU:        iface.MTU,
			IsUp:       iface.Flags&net.FlagUp != 0,
			IsLoopback: iface.Flags&net.FlagLoopback != 0,
		}

		// Get MAC address (HardwareAddr is a field, not a method)
		if len(iface.HardwareAddr) > 0 {
			ni.MAC = iface.HardwareAddr.String()
		}

		// Get IP addresses
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					ni.IP4 = v.IP.String()
					ni.Type = GetInterfaceType(iface.Name, v.IP)
				} else if v.IP.To16() != nil {
					ni.IP6 = v.IP.String()
				}
			}
		}

		interfaces = append(interfaces, ni)
	}

	return interfaces, nil
}

// GetInterfaceType determines the interface type based on name and IP
func GetInterfaceType(name string, ip net.IP) string {
	nameLower := strings.ToLower(name)
	
	// Check for known interface types
	if strings.HasPrefix(nameLower, "lo") || nameLower == "loopback" {
		return "loopback"
	}
	if strings.HasPrefix(nameLower, "wl") || strings.HasPrefix(nameLower, "wifi") || 
	   strings.HasPrefix(nameLower, "wlan") || strings.Contains(nameLower, "wireless") {
		return "wifi"
	}
	if strings.HasPrefix(nameLower, "eth") || strings.HasPrefix(nameLower, "en") || 
	   nameLower == "adapter" || strings.Contains(nameLower, "ethernet") {
		return "ethernet"
	}
	if strings.HasPrefix(nameLower, "tun") || strings.HasPrefix(nameLower, "tap") || 
	   strings.HasPrefix(nameLower, "ppp") || strings.HasPrefix(nameLower, "wg") ||
	   strings.HasPrefix(nameLower, "utun") || strings.HasPrefix(nameLower, "ipsec") {
		return "vpn"
	}
	if strings.HasPrefix(nameLower, "vmnet") || strings.HasPrefix(nameLower, "virtualbox") ||
	   strings.Contains(nameLower, "docker") || strings.HasPrefix(nameLower, "br-") ||
	   strings.HasPrefix(nameLower, "veth") {
		return "vm"
	}
	
	// Check IP ranges for common network types
	// Docker
	if ip.Equal(net.ParseIP("172.17.0.1")) || isInSubnet(ip, "172.17.0.0/16") {
		return "docker"
	}
	// VMware
	if isInSubnet(ip, "192.168.56.0/24") || isInSubnet(ip, "192.168.127.0/24") {
		return "vmware"
	}
	
	return "unknown"
}

func isInSubnet(ip net.IP, subnet string) bool {
	_, n, _ := net.ParseCIDR(subnet)
	return n != nil && n.Contains(ip)
}

// GetLoopbackInterface returns the loopback interface
func GetLoopbackInterface() (*types.NetworkInterface, error) {
	iface, err := net.InterfaceByName("lo")
	if err != nil {
		names := []string{"lo0", "lo"}
		for _, name := range names {
			if iface, err = net.InterfaceByName(name); err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("loopback interface not found")
		}
	}

	addrs, _ := iface.Addrs()
	ni := &types.NetworkInterface{
		Name:       iface.Name,
		Index:      iface.Index,
		MTU:        iface.MTU,
		IsUp:       iface.Flags&net.FlagUp != 0,
		IsLoopback: true,
		Type:       "loopback",
	}

	for _, addr := range addrs {
		if v, ok := addr.(*net.IPNet); ok {
			if v.IP.To4() != nil {
				ni.IP4 = v.IP.String()
			}
		}
	}

	return ni, nil
}
