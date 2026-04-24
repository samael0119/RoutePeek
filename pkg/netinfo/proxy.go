package netinfo

import (
	"os"
	"runtime"
	"strings"

	"github.com/routepeek/pkg/types"
)

// DetectProxy detects system proxy settings
func DetectProxy() *types.ProxyConfig {
	cfg := &types.ProxyConfig{}

	// Check environment variables
	for _, env := range ProxyEnvVars {
		if val := os.Getenv(env); val != "" {
			cfg.HasProxy = true
			envUpper := strings.ToUpper(env)
			if envUpper == "HTTP_PROXY" {
				cfg.HTTPProxy = val
			} else if envUpper == "HTTPS_PROXY" {
				cfg.HTTPSProxy = val
			} else if envUpper == "NO_PROXY" {
				cfg.NoProxy = val
			}
		}
	}

	// On Windows, also check WinHTTP proxy settings
	if runtime.GOOS == "windows" {
		if winProxy := getWindowsProxy(); winProxy != "" {
			cfg.HasProxy = true
			if cfg.HTTPProxy == "" {
				cfg.HTTPProxy = winProxy
			}
		}
	}

	return cfg
}

func getWindowsProxy() string {
	// Use netsh to get WinHTTP proxy (simplified)
	return ""
}
