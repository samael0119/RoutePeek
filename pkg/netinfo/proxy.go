package netinfo

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"

	"github.com/samael0119/RoutePeek/pkg/types"
)

// DetectProxy detects system proxy settings
func DetectProxy() *types.ProxyConfig {
	cfg := &types.ProxyConfig{}

	// Check environment variables
	for _, env := range ProxyEnvVars {
		if val := os.Getenv(env); val != "" {
			envUpper := strings.ToUpper(env)
			if envUpper == "HTTP_PROXY" {
				cfg.HasProxy = true
				cfg.HTTPProxy = val
			} else if envUpper == "HTTPS_PROXY" {
				cfg.HasProxy = true
				cfg.HTTPSProxy = val
			} else if envUpper == "ALL_PROXY" {
				cfg.HasProxy = true
				if cfg.HTTPProxy == "" {
					cfg.HTTPProxy = val
				}
				if cfg.HTTPSProxy == "" {
					cfg.HTTPSProxy = val
				}
			} else if envUpper == "NO_PROXY" {
				cfg.NoProxy = val
			}
		}
	}

	if runtime.GOOS == "windows" {
		mergeProxyConfig(cfg, getWindowsWinINetProxy())
		mergeProxyConfig(cfg, getWindowsWinHTTPProxy())
	}

	return cfg
}

func getWindowsWinINetProxy() *types.ProxyConfig {
	cmd := `Get-ItemProperty 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings' | Select-Object ProxyEnable,ProxyServer,AutoConfigURL | ConvertTo-Json -Compress`
	output, err := commandRunner("powershell", "-NoProfile", "-Command", cmd)
	if err != nil {
		return nil
	}
	return parseWindowsWinINetProxy(string(output))
}

func getWindowsWinHTTPProxy() *types.ProxyConfig {
	output, err := commandRunner("netsh", "winhttp", "show", "proxy")
	if err != nil {
		return nil
	}
	return parseWindowsWinHTTPProxy(string(output))
}

func parseWindowsWinINetProxy(raw string) *types.ProxyConfig {
	var payload struct {
		ProxyEnable   int    `json:"ProxyEnable"`
		ProxyServer   string `json:"ProxyServer"`
		AutoConfigURL string `json:"AutoConfigURL"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return nil
	}

	cfg := parseProxyServer(payload.ProxyServer, payload.ProxyEnable != 0)
	if cfg == nil {
		cfg = &types.ProxyConfig{}
	}
	if strings.TrimSpace(payload.AutoConfigURL) != "" {
		cfg.HasProxy = true
		if cfg.HTTPProxy == "" {
			cfg.HTTPProxy = strings.TrimSpace(payload.AutoConfigURL)
		}
	}
	if !cfg.HasProxy {
		return nil
	}
	return cfg
}

func parseWindowsWinHTTPProxy(output string) *types.ProxyConfig {
	cfg := &types.ProxyConfig{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "proxy server(s)":
			mergeProxyConfig(cfg, parseProxyServer(value, true))
		case "bypass list":
			cfg.NoProxy = value
		}
	}
	if !cfg.HasProxy {
		return nil
	}
	return cfg
}

func parseProxyServer(value string, enabled bool) *types.ProxyConfig {
	value = strings.TrimSpace(value)
	if !enabled || value == "" {
		return nil
	}

	cfg := &types.ProxyConfig{HasProxy: true}
	if !strings.Contains(value, "=") {
		cfg.HTTPProxy = value
		cfg.HTTPSProxy = value
		return cfg
	}

	for _, part := range strings.Split(value, ";") {
		key, proxy, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "http":
			cfg.HTTPProxy = strings.TrimSpace(proxy)
		case "https":
			cfg.HTTPSProxy = strings.TrimSpace(proxy)
		}
	}
	if cfg.HTTPProxy == "" && cfg.HTTPSProxy != "" {
		cfg.HTTPProxy = cfg.HTTPSProxy
	}
	if cfg.HTTPSProxy == "" && cfg.HTTPProxy != "" {
		cfg.HTTPSProxy = cfg.HTTPProxy
	}
	if cfg.HTTPProxy == "" && cfg.HTTPSProxy == "" {
		return nil
	}
	return cfg
}

func mergeProxyConfig(dst *types.ProxyConfig, src *types.ProxyConfig) {
	if dst == nil || src == nil || !src.HasProxy {
		return
	}
	dst.HasProxy = true
	if dst.HTTPProxy == "" {
		dst.HTTPProxy = src.HTTPProxy
	}
	if dst.HTTPSProxy == "" {
		dst.HTTPSProxy = src.HTTPSProxy
	}
	if dst.NoProxy == "" {
		dst.NoProxy = src.NoProxy
	}
}
