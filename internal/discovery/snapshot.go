package discovery

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/samael0119/RoutePeek/internal/diagnostic"
	"github.com/samael0119/RoutePeek/pkg/netinfo"
	"github.com/samael0119/RoutePeek/pkg/types"
)

var (
	getNetworkInterfaces = netinfo.GetNetworkInterfaces
	getLoopbackInterface = netinfo.GetLoopbackInterface
	getRoutes            = netinfo.GetRoutes
	getDefaultGateway    = netinfo.GetDefaultGateway
	getDNSConfig         = netinfo.GetDNSConfig
	detectProxy          = netinfo.DetectProxy
	detectVPN            = netinfo.DetectVPN
	detectVMNetworks     = netinfo.DetectVMNetworks
	fetchPublicIP        = getPublicIP
)

// GetNetworkSnapshot returns a complete snapshot of network state
func GetNetworkSnapshot() (*types.NetworkSnapshot, error) {
	snapshot := &types.NetworkSnapshot{
		Timestamp: time.Now(),
	}

	// Get network interfaces
	if interfaces, err := getNetworkInterfaces(); err == nil {
		snapshot.Interfaces = interfaces
	}

	// Get loopback separately
	if lo, err := getLoopbackInterface(); err == nil {
		snapshot.Interfaces = append(snapshot.Interfaces, *lo)
	}

	// Get routes
	if routes, err := getRoutes(); err == nil {
		snapshot.Routes = routes
	}

	// Get default gateway
	if gw, err := getDefaultGateway(); err == nil {
		snapshot.DefaultGateway = gw
	}

	// Get DNS config
	if dns, err := getDNSConfig(); err == nil {
		snapshot.DNS = *dns
	}

	// Get proxy config
	if proxy := detectProxy(); proxy != nil {
		snapshot.Proxy = *proxy
	}

	// Get VPN info
	snapshot.VPN = detectVPN()

	// Get VM networks
	snapshot.VMNetworks = detectVMNetworks()

	// Get public IP (optional, may fail)
	if pubIP, err := fetchPublicIP(); err == nil {
		snapshot.PublicIP = pubIP
	}

	normalizeSnapshotCollections(snapshot)

	return snapshot, nil
}

func normalizeSnapshotCollections(snapshot *types.NetworkSnapshot) {
	if snapshot.Interfaces == nil {
		snapshot.Interfaces = []types.NetworkInterface{}
	}
	if snapshot.Routes == nil {
		snapshot.Routes = []types.RouteEntry{}
	}
	if snapshot.DNS.Servers == nil {
		snapshot.DNS.Servers = []string{}
	}
	if snapshot.DNS.Search == nil {
		snapshot.DNS.Search = []string{}
	}
	if snapshot.VMNetworks == nil {
		snapshot.VMNetworks = []types.VMNetwork{}
	}
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
	client := &http.Client{Timeout: 2 * time.Second}
	services := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		if ip := strings.TrimSpace(string(body)); ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("could not determine public IP")
}
