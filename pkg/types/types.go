package types

import (
	"time"
)

// NetworkInterface represents a network interface on the machine
type NetworkInterface struct {
	Name       string `json:"name"`
	Index      int    `json:"index"`
	MTU        int    `json:"mtu"`
	MAC        string `json:"mac"`
	IP4        string `json:"ip4"`
	IP6        string `json:"ip6"`
	IsUp       bool   `json:"is_up"`
	IsLoopback bool   `json:"is_loopback"`
	Type       string `json:"type"`
}

// RouteEntry represents a single route table entry
type RouteEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	Table       string `json:"table"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	Servers  []string `json:"servers"`
	Search   []string `json:"search"`
	Port     int      `json:"port"`
	IsSecure bool     `json:"is_secure"`
}

// ProxyConfig represents proxy settings
type ProxyConfig struct {
	HTTPProxy  string `json:"http_proxy"`
	HTTPSProxy string `json:"https_proxy"`
	NoProxy    string `json:"no_proxy"`
	HasProxy   bool   `json:"has_proxy"`
}

// VPNInfo represents VPN connection status
type VPNInfo struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	ServerIP  string `json:"server_ip"`
	ClientIP  string `json:"client_ip"`
	Protocol  string `json:"protocol"`
	Interface string `json:"interface"`
	IsGlobal  bool   `json:"is_global"`
}

// VMNetwork represents virtual machine network info
type VMNetwork struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	DHCP    bool   `json:"dhcp"`
	HostIP  string `json:"host_ip"`
}

// NetworkSnapshot represents a complete snapshot of network state
type NetworkSnapshot struct {
	Timestamp      time.Time          `json:"timestamp"`
	Interfaces     []NetworkInterface `json:"interfaces"`
	Routes         []RouteEntry      `json:"routes"`
	DNS            DNSConfig         `json:"dns"`
	Proxy          ProxyConfig       `json:"proxy"`
	VPN            *VPNInfo          `json:"vpn,omitempty"`
	VMNetworks     []VMNetwork       `json:"vm_networks"`
	DefaultGateway string             `json:"default_gateway"`
	PublicIP       string             `json:"public_ip"`
}

// DiagnosisResult represents a single diagnostic finding
type DiagnosisResult struct {
	Severity   string `json:"severity"`
	Code       string `json:"code"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// DiagnosisReport contains all diagnostic findings
type DiagnosisReport struct {
	Timestamp time.Time         `json:"timestamp"`
	Findings  []DiagnosisResult `json:"findings"`
	Summary   string            `json:"summary"`
}
