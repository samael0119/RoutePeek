package discovery

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/routepeek/internal/diagnostic"
	"github.com/routepeek/pkg/netinfo"
	"github.com/routepeek/pkg/types"
)

// GetNetworkSnapshot returns a complete snapshot of network state
func GetNetworkSnapshot() (*types.NetworkSnapshot, error) {
	snapshot := &types.NetworkSnapshot{
		Timestamp: time.Now(),
	}

	// Get network interfaces
	interfaces, err := netinfo.GetNetworkInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}
	snapshot.Interfaces = interfaces

	// Get loopback separately
	if lo, err := netinfo.GetLoopbackInterface(); err == nil {
		snapshot.Interfaces = append(snapshot.Interfaces, *lo)
	}

	// Get routes
	routes, err := netinfo.GetRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}
	snapshot.Routes = routes

	// Get default gateway
	if gw, err := netinfo.GetDefaultGateway(); err == nil {
		snapshot.DefaultGateway = gw
	}

	// Get DNS config
	dns, err := netinfo.GetDNSConfig()
	if err == nil {
		snapshot.DNS = *dns
	}

	// Get proxy config
	snapshot.Proxy = *netinfo.DetectProxy()

	// Get VPN info
	snapshot.VPN = netinfo.DetectVPN()

	// Get VM networks
	snapshot.VMNetworks = netinfo.DetectVMNetworks()

	// Get public IP (optional, may fail)
	if pubIP, err := getPublicIP(); err == nil {
		snapshot.PublicIP = pubIP
	}

	return snapshot, nil
}

// GetInterfaces returns all network interfaces
func GetInterfaces() ([]types.NetworkInterface, error) {
	return netinfo.GetNetworkInterfaces()
}

// GetRoutes returns routing table entries
func GetRoutes() ([]types.RouteEntry, error) {
	return netinfo.GetRoutes()
}

// RunDiagnostics runs diagnostics on the network snapshot
func RunDiagnostics(snapshot *types.NetworkSnapshot) *types.DiagnosisReport {
	return diagnostic.RunDiagnostics(snapshot)
}

// getPublicIP retrieves the public/external IP address
func getPublicIP() (string, error) {
	services := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, service := range services {
		resp, err := http.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		return string(body), nil
	}

	return "", fmt.Errorf("could not determine public IP")
}
