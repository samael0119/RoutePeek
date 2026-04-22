package netinfo

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/netscope/pkg/types"
)

// GetRoutes returns the routing table entries
func GetRoutes() ([]types.RouteEntry, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxRoutes()
	case "darwin":
		return getMacRoutes()
	case "windows":
		return getWindowsRoutes()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// GetDefaultGateway returns the default gateway
func GetDefaultGateway() (string, error) {
	routes, err := GetRoutes()
	if err != nil {
		return "", err
	}

	for _, route := range routes {
		if route.Destination == "0.0.0.0/0" || route.Destination == "::/0" {
			return route.Gateway, nil
		}
	}

	return "", fmt.Errorf("no default gateway found")
}

// GetDefaultInterface returns the interface used for default route
func GetDefaultInterface() (string, error) {
	routes, err := GetRoutes()
	if err != nil {
		return "", err
	}

	for _, route := range routes {
		if route.Destination == "0.0.0.0/0" {
			return route.Interface, nil
		}
	}

	return "", fmt.Errorf("no default route found")
}

// Linux route parsing
func getLinuxRoutes() ([]types.RouteEntry, error) {
	cmd := exec.Command("ip", "route", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'ip route': %w", err)
	}

	var routes []types.RouteEntry
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		entry := types.RouteEntry{
			Table: "main",
		}

		// First field is usually destination
		if parts[0] == "default" {
			entry.Destination = "0.0.0.0/0"
		} else if strings.Contains(parts[0], "/") {
			entry.Destination = parts[0]
		}

		// Parse rest of the line
		for i := 1; i < len(parts); i++ {
			switch parts[i] {
			case "via":
				if i+1 < len(parts) {
					entry.Gateway = parts[i+1]
				}
			case "dev":
				if i+1 < len(parts) {
					entry.Interface = parts[i+1]
				}
			case "metric":
				if i+1 < len(parts) {
					if m, err := strconv.Atoi(parts[i+1]); err == nil {
						entry.Metric = m
					}
				}
			}
		}

		if entry.Destination != "" && entry.Interface != "" {
			routes = append(routes, entry)
		}
	}

	return routes, nil
}

// macOS route parsing
func getMacRoutes() ([]types.RouteEntry, error) {
	cmd := exec.Command("netstat", "-rn")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'netstat -rn': %w", err)
	}

	var routes []types.RouteEntry
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	inIPv4Section := false

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header lines
		if strings.HasPrefix(line, "Internet") || strings.HasPrefix(line, "Internet6") {
			inIPv4Section = strings.HasPrefix(line, "Internet") && !strings.HasPrefix(line, "Internet6")
			continue
		}

		if !inIPv4Section {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		entry := types.RouteEntry{
			Destination: parts[0],
			Gateway:     parts[1],
			Interface:   parts[3],
			Table:       "main",
		}

		if len(parts) >= 7 {
			if m, err := strconv.Atoi(parts[6]); err == nil {
				entry.Metric = m
			}
		}

		if entry.Destination != "" {
			routes = append(routes, entry)
		}
	}

	return routes, nil
}

// Windows route parsing
func getWindowsRoutes() ([]types.RouteEntry, error) {
	cmd := exec.Command("route", "print", "0.0.0.0")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'route print': %w", err)
	}

	var routes []types.RouteEntry
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	capture := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Start capturing when we see IPv4 Route table header
		if strings.Contains(line, "IPv4 Route Table") || strings.Contains(line, "IPv4") {
			capture = true
			continue
		}

		if strings.Contains(line, "IPv6 Route Table") || strings.Contains(line, "IPv6") {
			capture = false
			continue
		}

		if !capture {
			continue
		}

		// Parse route entries (typical format: Network Destination  Netmask  Gateway  Interface  Metric)
		parts := strings.Fields(line)
		if len(parts) >= 5 {
			entry := types.RouteEntry{
				Destination: parts[0],
				Gateway:     parts[2],
				Interface:   parts[3],
				Table:       "main",
			}

			if len(parts) >= 6 {
				if m, err := strconv.Atoi(parts[5]); err == nil {
					entry.Metric = m
				}
			}

			if entry.Destination != "" && entry.Destination != "Gateway" {
				routes = append(routes, entry)
			}
		}
	}

	return routes, nil
}

// GetInterfaceRoutes returns routes associated with a specific interface
func GetInterfaceRoutes(ifaceName string) ([]types.RouteEntry, error) {
	routes, err := GetRoutes()
	if err != nil {
		return nil, err
	}

	var ifaceRoutes []types.RouteEntry
	for _, route := range routes {
		if route.Interface == ifaceName {
			ifaceRoutes = append(ifaceRoutes, route)
		}
	}

	return ifaceRoutes, nil
}

// IsVPNGlobal checks if VPN routes all traffic (0.0.0.0/0 through VPN)
func IsVPNGlobal(vpnInterface string, routes []types.RouteEntry) bool {
	for _, route := range routes {
		if route.Interface == vpnInterface {
			if route.Destination == "0.0.0.0/0" {
				return true
			}
		}
	}
	return false
}

// GetVMSubnetRoutes returns routes that go to VM networks
func GetVMSubnetRoutes(routes []types.RouteEntry) []types.RouteEntry {
	var vmRoutes []types.RouteEntry

	for _, route := range routes {
		// Check if destination is a known VM subnet
		_, dstNet, _ := net.ParseCIDR(route.Destination)
		if dstNet == nil {
			continue
		}

		for _, vmSubnet := range append(CommonVMwareSubnets, append(CommonVirtualBoxSubnets, CommonDockerSubnets...)...) {
			_, vmNet, _ := net.ParseCIDR(vmSubnet)
			if vmNet != nil && vmNet.IP.Equal(dstNet.IP) {
				vmRoutes = append(vmRoutes, route)
				break
			}
		}
	}

	return vmRoutes
}
