package netinfo

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/routepeek/pkg/types"
)

// GetDNSConfig returns DNS configuration
func GetDNSConfig() (*types.DNSConfig, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxDNS()
	case "darwin":
		return getMacDNS()
	case "windows":
		return getWindowsDNS()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Well-known public DNS servers
var PublicDNS = map[string]bool{
	"8.8.8.8":            true,
	"8.8.4.4":            true,
	"1.1.1.1":            true,
	"1.0.0.1":            true,
	"9.9.9.9":            true,
	"208.67.222.222":     true,
	"208.67.220.220":     true,
	"114.114.114.114":    true,
	"114.114.115.115":    true,
	"223.5.5.5":          true,
	"223.6.6.6":          true,
}

// IsPublicDNS checks if the DNS server is a known public DNS
func IsPublicDNS(ip string) bool {
	return PublicDNS[ip]
}

// IsSecureDNS checks if DNS uses DoH/DoT
func IsSecureDNS(server string) bool {
	securePrefixes := []string{"https://", "tls://", "bootp://"}
	for _, prefix := range securePrefixes {
		if strings.HasPrefix(server, prefix) {
			return true
		}
	}
	return false
}

func getLinuxDNS() (*types.DNSConfig, error) {
	cfg := &types.DNSConfig{Port: 53}

	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/resolv.conf: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "nameserver":
			cfg.Servers = append(cfg.Servers, parts[1])
			if IsSecureDNS(parts[1]) {
				cfg.IsSecure = true
			}
		case "search":
			cfg.Search = append(cfg.Search, parts[1:]...)
		}
	}

	return cfg, nil
}

func getMacDNS() (*types.DNSConfig, error) {
	cfg := &types.DNSConfig{Port: 53}

	cmd := exec.Command("scutil", "--dns")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS config: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				cfg.Servers = append(cfg.Servers, parts[1])
				if IsSecureDNS(parts[1]) {
					cfg.IsSecure = true
				}
			}
		}
	}

	return cfg, nil
}

func getWindowsDNS() (*types.DNSConfig, error) {
	cfg := &types.DNSConfig{Port: 53}

	cmd := exec.Command("ipconfig", "/all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS config: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "DNS Servers") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if net.ParseIP(part) != nil {
					cfg.Servers = append(cfg.Servers, part)
					if IsSecureDNS(part) {
						cfg.IsSecure = true
					}
				}
			}
		}
	}

	return cfg, nil
}
