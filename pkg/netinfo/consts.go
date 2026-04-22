package netinfo

// Proxy environment variables
var ProxyEnvVars = []string{
	"HTTP_PROXY", "http_proxy",
	"HTTPS_PROXY", "https_proxy",
	"ALL_PROXY", "all_proxy",
	"NO_PROXY", "no_proxy",
}

// VPN interface name prefixes
var VPNInterfacePrefixes = []string{
	"tun", "tap", "ppp", "vpnc", "vpn",
	"wg", "utun", "ipsec",
}

// Common VMware subnets
var CommonVMwareSubnets = []string{
	"192.168.56.0/24",  // VMnet8 (NAT)
	"192.168.127.0/24", // VMnet1 (host-only)
	"192.168.217.0/24", // VMnet8 (NAT) - some versions
	"172.16.0.0/16",    // Hyper-V default switch
}

// Common VirtualBox subnets
var CommonVirtualBoxSubnets = []string{
	"192.168.56.0/24",  // VirtualBox NAT
	"10.0.2.0/24",      // VirtualBox default NAT
	"192.168.59.0/24",  // VirtualBox host-only
}

// Common Docker subnets
var CommonDockerSubnets = []string{
	"172.17.0.0/16",   // Docker default bridge
	"172.18.0.0/16",   // Docker user-defined bridge
	"172.19.0.0/16",
	"172.20.0.0/16",
	"172.21.0.0/16",
	"172.22.0.0/16",
}

// KnownVMSubnets is a combined list of known VM subnets
var KnownVMSubnets = []struct {
	Name   string
	Type   string
	Subnet string
}{
	{"VMware VMnet8", "nat", "192.168.56.0/24"},
	{"VMware VMnet1", "host-only", "192.168.127.0/24"},
	{"VirtualBox NAT", "nat", "192.168.56.0/24"},
	{"Docker bridge", "bridge", "172.17.0.0/16"},
	{"Hyper-V", "switch", "172.16.0.0/16"},
}
