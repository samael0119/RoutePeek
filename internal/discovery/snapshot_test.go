package discovery

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestGetNetworkSnapshotToleratesInterfaceCollectionFailure(t *testing.T) {
	origInterfaces := getNetworkInterfaces
	origLoopback := getLoopbackInterface
	origRoutes := getRoutes
	origDefaultGateway := getDefaultGateway
	origDNS := getDNSConfig
	origProxy := detectProxy
	origVPN := detectVPN
	origVMNetworks := detectVMNetworks
	origPublicIP := fetchPublicIP
	t.Cleanup(func() {
		getNetworkInterfaces = origInterfaces
		getLoopbackInterface = origLoopback
		getRoutes = origRoutes
		getDefaultGateway = origDefaultGateway
		getDNSConfig = origDNS
		detectProxy = origProxy
		detectVPN = origVPN
		detectVMNetworks = origVMNetworks
		fetchPublicIP = origPublicIP
	})

	getNetworkInterfaces = func() ([]types.NetworkInterface, error) {
		return nil, errors.New("interface command unavailable")
	}
	getLoopbackInterface = func() (*types.NetworkInterface, error) {
		return nil, errors.New("loopback unavailable")
	}
	getRoutes = func() ([]types.RouteEntry, error) {
		return []types.RouteEntry{{Destination: "0.0.0.0/0", Gateway: "192.168.1.1"}}, nil
	}
	getDefaultGateway = func() (string, error) {
		return "192.168.1.1", nil
	}
	getDNSConfig = func() (*types.DNSConfig, error) {
		return &types.DNSConfig{Servers: []string{"1.1.1.1"}}, nil
	}
	detectProxy = func() *types.ProxyConfig {
		return &types.ProxyConfig{}
	}
	detectVPN = func() *types.VPNInfo {
		return nil
	}
	detectVMNetworks = func() []types.VMNetwork {
		return nil
	}
	fetchPublicIP = func() (string, error) {
		return "203.0.113.10", nil
	}

	snapshot, err := GetNetworkSnapshot()

	if err != nil {
		t.Fatalf("expected snapshot despite interface failure, got error %v", err)
	}
	if snapshot == nil {
		t.Fatalf("expected snapshot")
	}
	if len(snapshot.Interfaces) != 0 {
		t.Fatalf("expected empty interfaces after collection failure, got %#v", snapshot.Interfaces)
	}
	if snapshot.DefaultGateway != "192.168.1.1" {
		t.Fatalf("expected remaining collectors to run, got %#v", snapshot)
	}
}

func TestGetNetworkSnapshotToleratesRouteCollectionFailure(t *testing.T) {
	origInterfaces := getNetworkInterfaces
	origLoopback := getLoopbackInterface
	origRoutes := getRoutes
	origDefaultGateway := getDefaultGateway
	origDNS := getDNSConfig
	origProxy := detectProxy
	origVPN := detectVPN
	origVMNetworks := detectVMNetworks
	origPublicIP := fetchPublicIP
	t.Cleanup(func() {
		getNetworkInterfaces = origInterfaces
		getLoopbackInterface = origLoopback
		getRoutes = origRoutes
		getDefaultGateway = origDefaultGateway
		getDNSConfig = origDNS
		detectProxy = origProxy
		detectVPN = origVPN
		detectVMNetworks = origVMNetworks
		fetchPublicIP = origPublicIP
	})

	getNetworkInterfaces = func() ([]types.NetworkInterface, error) {
		return []types.NetworkInterface{{Name: "eth0", IP4: "192.168.1.20", IsUp: true}}, nil
	}
	getLoopbackInterface = func() (*types.NetworkInterface, error) {
		return nil, errors.New("loopback unavailable")
	}
	getRoutes = func() ([]types.RouteEntry, error) {
		return nil, errors.New("route command unavailable")
	}
	getDefaultGateway = func() (string, error) {
		return "", errors.New("gateway unavailable")
	}
	getDNSConfig = func() (*types.DNSConfig, error) {
		return &types.DNSConfig{Servers: []string{"1.1.1.1"}}, nil
	}
	detectProxy = func() *types.ProxyConfig {
		return &types.ProxyConfig{}
	}
	detectVPN = func() *types.VPNInfo {
		return nil
	}
	detectVMNetworks = func() []types.VMNetwork {
		return nil
	}
	fetchPublicIP = func() (string, error) {
		return "203.0.113.10", nil
	}

	snapshot, err := GetNetworkSnapshot()

	if err != nil {
		t.Fatalf("expected snapshot despite route failure, got error %v", err)
	}
	if snapshot == nil {
		t.Fatalf("expected snapshot")
	}
	if len(snapshot.Routes) != 0 {
		t.Fatalf("expected empty routes after collection failure, got %#v", snapshot.Routes)
	}
	if len(snapshot.Interfaces) != 1 {
		t.Fatalf("expected interface data to be preserved, got %#v", snapshot.Interfaces)
	}
}

func TestGetNetworkSnapshotEncodesEmptyCollectionFieldsAsArrays(t *testing.T) {
	origInterfaces := getNetworkInterfaces
	origLoopback := getLoopbackInterface
	origRoutes := getRoutes
	origDefaultGateway := getDefaultGateway
	origDNS := getDNSConfig
	origProxy := detectProxy
	origVPN := detectVPN
	origVMNetworks := detectVMNetworks
	origPublicIP := fetchPublicIP
	t.Cleanup(func() {
		getNetworkInterfaces = origInterfaces
		getLoopbackInterface = origLoopback
		getRoutes = origRoutes
		getDefaultGateway = origDefaultGateway
		getDNSConfig = origDNS
		detectProxy = origProxy
		detectVPN = origVPN
		detectVMNetworks = origVMNetworks
		fetchPublicIP = origPublicIP
	})

	getNetworkInterfaces = func() ([]types.NetworkInterface, error) {
		return nil, errors.New("interface command unavailable")
	}
	getLoopbackInterface = func() (*types.NetworkInterface, error) {
		return nil, errors.New("loopback unavailable")
	}
	getRoutes = func() ([]types.RouteEntry, error) {
		return nil, errors.New("route command unavailable")
	}
	getDefaultGateway = func() (string, error) {
		return "", errors.New("gateway unavailable")
	}
	getDNSConfig = func() (*types.DNSConfig, error) {
		return &types.DNSConfig{}, nil
	}
	detectProxy = func() *types.ProxyConfig {
		return nil
	}
	detectVPN = func() *types.VPNInfo {
		return nil
	}
	detectVMNetworks = func() []types.VMNetwork {
		return nil
	}
	fetchPublicIP = func() (string, error) {
		return "", errors.New("public ip unavailable")
	}

	snapshot, err := GetNetworkSnapshot()
	if err != nil {
		t.Fatalf("expected snapshot despite collection failures, got error %v", err)
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	body := string(payload)

	for _, field := range []string{"interfaces", "routes", "servers", "search", "vm_networks"} {
		if strings.Contains(body, `"`+field+`":null`) {
			t.Fatalf("expected %s to encode as an empty array, got %s", field, body)
		}
	}
}
