# Windows Proxy/VPN Detection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make RoutePeek reliably detect enabled Windows system proxy and common Windows VPN/tunnel adapters.

**Architecture:** Keep detection in `pkg/netinfo` and keep UI/API types stable. Add small, testable parsers around Windows command output and PowerShell JSON, then wire them into `DetectProxy` and `DetectVPN`. Avoid guessing from localized display strings when a stable registry/PowerShell field is available.

**Tech Stack:** Go 1.21, existing `commandRunner`, Windows `netsh`, Windows PowerShell registry/network adapter commands, existing Vue UI.

---

## File Structure

- Modify: `pkg/netinfo/proxy.go`
  - Implement Windows proxy detection from WinINet registry and WinHTTP output.
  - Keep environment variable detection as the first source.
- Create: `pkg/netinfo/proxy_test.go`
  - Cover WinINet JSON, WinHTTP output, direct-access output, and `ALL_PROXY`.
- Modify: `pkg/netinfo/vpn.go`
  - Add injectable interface providers and Windows adapter name heuristics.
  - Add Windows route-based fallback that can match route interface IPs to adapter addresses.
- Create: `pkg/netinfo/vpn_test.go`
  - Cover common Windows VPN adapter names and route fallback.
- Modify: `pkg/netinfo/consts.go`
  - Add common Windows VPN/tunnel keywords without weakening Unix prefix checks.
- Optional Modify: `internal/discovery/snapshot_test.go`
  - Add a snapshot-level assertion that proxy/VPN values are preserved.

---

## Task 1: Proxy Detection Evidence and Parser Tests

**Files:**
- Create: `pkg/netinfo/proxy_test.go`
- Modify later: `pkg/netinfo/proxy.go`

- [ ] **Step 1: Write failing tests for Windows proxy parsers**

Add `pkg/netinfo/proxy_test.go`:

```go
package netinfo

import "testing"

func TestParseWindowsWinINetProxyEnabled(t *testing.T) {
	got := parseWindowsWinINetProxy(`{"ProxyEnable":1,"ProxyServer":"127.0.0.1:7890","AutoConfigURL":""}`)
	if got == nil || !got.HasProxy || got.HTTPProxy != "127.0.0.1:7890" || got.HTTPSProxy != "127.0.0.1:7890" {
		t.Fatalf("unexpected proxy config %#v", got)
	}
}

func TestParseWindowsWinINetPACEnabled(t *testing.T) {
	got := parseWindowsWinINetProxy(`{"ProxyEnable":0,"ProxyServer":"","AutoConfigURL":"http://127.0.0.1:7890/proxy.pac"}`)
	if got == nil || !got.HasProxy || got.HTTPProxy != "http://127.0.0.1:7890/proxy.pac" {
		t.Fatalf("unexpected pac proxy config %#v", got)
	}
}

func TestParseWindowsWinHTTPProxy(t *testing.T) {
	output := "Current WinHTTP proxy settings:\r\n\r\n    Proxy Server(s) : 127.0.0.1:7890\r\n    Bypass List     : localhost;*.local\r\n"
	got := parseWindowsWinHTTPProxy(output)
	if got == nil || !got.HasProxy || got.HTTPProxy != "127.0.0.1:7890" || got.NoProxy != "localhost;*.local" {
		t.Fatalf("unexpected winhttp proxy config %#v", got)
	}
}

func TestParseWindowsWinHTTPDirectAccess(t *testing.T) {
	output := "Current WinHTTP proxy settings:\r\n\r\n    Direct access (no proxy server).\r\n"
	if got := parseWindowsWinHTTPProxy(output); got != nil && got.HasProxy {
		t.Fatalf("expected no proxy, got %#v", got)
	}
}
```

- [ ] **Step 2: Run proxy parser tests and verify they fail**

Run:

```bash
go test ./pkg/netinfo -run 'TestParseWindows.*Proxy' -count=1
```

Expected: build fails because `parseWindowsWinINetProxy` and `parseWindowsWinHTTPProxy` do not exist.

---

## Task 2: Implement Windows Proxy Detection

**Files:**
- Modify: `pkg/netinfo/proxy.go`
- Test: `pkg/netinfo/proxy_test.go`

- [ ] **Step 1: Add parsers and command collectors**

Implement these helpers in `pkg/netinfo/proxy.go`:

```go
func parseWindowsWinINetProxy(raw string) *types.ProxyConfig {
	var payload struct {
		ProxyEnable  int    `json:"ProxyEnable"`
		ProxyServer  string `json:"ProxyServer"`
		AutoConfigURL string `json:"AutoConfigURL"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return nil
	}
	cfg := &types.ProxyConfig{}
	if strings.TrimSpace(payload.ProxyServer) != "" && payload.ProxyEnable != 0 {
		cfg.HasProxy = true
		cfg.HTTPProxy = payload.ProxyServer
		cfg.HTTPSProxy = payload.ProxyServer
	}
	if strings.TrimSpace(payload.AutoConfigURL) != "" {
		cfg.HasProxy = true
		if cfg.HTTPProxy == "" {
			cfg.HTTPProxy = payload.AutoConfigURL
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
			if value != "" {
				cfg.HasProxy = true
				cfg.HTTPProxy = value
				cfg.HTTPSProxy = value
			}
		case "bypass list":
			cfg.NoProxy = value
		}
	}
	if !cfg.HasProxy {
		return nil
	}
	return cfg
}
```

- [ ] **Step 2: Wire Windows collection into `DetectProxy`**

Use PowerShell for WinINet because it returns stable field names independent of display language:

```go
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
```

Merge results without overwriting environment values already set:

```go
func mergeProxyConfig(dst *types.ProxyConfig, src *types.ProxyConfig) {
	if src == nil || !src.HasProxy {
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
```

- [ ] **Step 3: Run proxy tests**

Run:

```bash
go test ./pkg/netinfo -run 'TestParseWindows.*Proxy' -count=1
```

Expected: PASS.

---

## Task 3: VPN Adapter Heuristic Tests

**Files:**
- Create: `pkg/netinfo/vpn_test.go`
- Modify later: `pkg/netinfo/vpn.go`

- [ ] **Step 1: Add tests for Windows adapter names**

Add `pkg/netinfo/vpn_test.go`:

```go
package netinfo

import "testing"

func TestLooksLikeVPNInterfaceMatchesWindowsVPNNames(t *testing.T) {
	names := []string{
		"WireGuard Tunnel",
		"Tailscale",
		"Cisco AnyConnect Secure Mobility Client Virtual Miniport Adapter",
		"Fortinet SSL VPN Virtual Ethernet Adapter",
		"TAP-Windows Adapter V9",
		"OpenVPN TAP-Windows6",
		"ZeroTier One Virtual Port",
		"Cloudflare WARP",
	}
	for _, name := range names {
		if !looksLikeVPNInterface(name) {
			t.Fatalf("expected %q to look like vpn", name)
		}
	}
}

func TestLooksLikeVPNInterfaceDoesNotMatchCommonPhysicalAdapters(t *testing.T) {
	names := []string{
		"Intel(R) Wi-Fi 6 AX201",
		"Realtek PCIe GbE Family Controller",
		"Bluetooth Device (Personal Area Network)",
		"Hyper-V Virtual Ethernet Adapter",
		"VMware Virtual Ethernet Adapter for VMnet8",
	}
	for _, name := range names {
		if looksLikeVPNInterface(name) {
			t.Fatalf("expected %q to not look like vpn", name)
		}
	}
}
```

- [ ] **Step 2: Run VPN name tests and verify they fail**

Run:

```bash
go test ./pkg/netinfo -run TestLooksLikeVPNInterface -count=1
```

Expected: build fails because `looksLikeVPNInterface` does not exist.

---

## Task 4: Implement VPN Adapter Heuristics

**Files:**
- Modify: `pkg/netinfo/consts.go`
- Modify: `pkg/netinfo/vpn.go`
- Test: `pkg/netinfo/vpn_test.go`

- [ ] **Step 1: Add Windows VPN keywords**

In `pkg/netinfo/consts.go`, add:

```go
var WindowsVPNInterfaceKeywords = []string{
	"vpn",
	"wireguard",
	"tailscale",
	"anyconnect",
	"fortinet",
	"globalprotect",
	"openvpn",
	"tap-windows",
	"zerotier",
	"warp",
}
```

- [ ] **Step 2: Add `looksLikeVPNInterface`**

In `pkg/netinfo/vpn.go`, add:

```go
func looksLikeVPNInterface(name string) bool {
	lower := strings.ToLower(name)
	for _, prefix := range VPNInterfacePrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	for _, keyword := range WindowsVPNInterfaceKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
```

Then replace the current prefix loop in `DetectVPN` with:

```go
if looksLikeVPNInterface(iface.Name) {
	// existing connected VPN assignment block
}
```

- [ ] **Step 3: Run VPN name tests**

Run:

```bash
go test ./pkg/netinfo -run TestLooksLikeVPNInterface -count=1
```

Expected: PASS.

---

## Task 5: Windows Route-Based VPN Fallback

**Files:**
- Modify: `pkg/netinfo/vpn.go`
- Test: `pkg/netinfo/vpn_test.go`

- [ ] **Step 1: Add route fallback tests around interface IP matching**

Add to `pkg/netinfo/vpn_test.go`:

```go
func TestVPNRouteFallbackMatchesRouteInterfaceIP(t *testing.T) {
	iface := types.NetworkInterface{Name: "WireGuard Tunnel", IP4: "10.8.0.2", IsUp: true, Type: "vpn"}
	routes := []types.RouteEntry{
		{Destination: "0.0.0.0/0", Gateway: "10.8.0.1", Interface: "10.8.0.2"},
	}
	got := vpnFromRoutes(routes, "192.168.1.1", []types.NetworkInterface{iface})
	if got == nil || got.Status != "connected" || got.Interface != "WireGuard Tunnel" {
		t.Fatalf("unexpected vpn from route fallback %#v", got)
	}
}
```

- [ ] **Step 2: Implement route fallback helper**

Add to `pkg/netinfo/vpn.go`:

```go
func vpnFromRoutes(routes []types.RouteEntry, defaultGw string, interfaces []types.NetworkInterface) *types.VPNInfo {
	for _, route := range routes {
		if route.Destination != "default" && route.Destination != "0.0.0.0/0" {
			continue
		}
		if route.Interface == "" || route.Gateway == "" || route.Gateway == defaultGw {
			continue
		}
		for _, iface := range interfaces {
			if iface.IP4 == route.Interface || iface.Name == route.Interface || looksLikeVPNInterface(iface.Name) {
				return &types.VPNInfo{
					Name:      iface.Name,
					Interface: iface.Name,
					Status:    "connected",
					ClientIP:  iface.IP4,
					Protocol:  guessVPNProtocol(net.ParseIP(iface.IP4)),
				}
			}
		}
	}
	return nil
}
```

Use `GetNetworkInterfaces()` inside `DetectVPN` to feed this helper after the direct `net.Interfaces()` pass.

- [ ] **Step 3: Run route fallback test**

Run:

```bash
go test ./pkg/netinfo -run TestVPNRouteFallbackMatchesRouteInterfaceIP -count=1
```

Expected: PASS.

---

## Task 6: End-to-End Verification and Release Build

**Files:**
- No new source files unless earlier tasks require small fixes.
- Release artifacts in `dist/`.

- [ ] **Step 1: Run full Go tests**

Run:

```bash
go test ./...
```

Expected: all packages pass.

- [ ] **Step 2: Run frontend tests**

Run:

```bash
cd web && npm test
```

Expected: all Vitest tests pass.

- [ ] **Step 3: Build all v0.2.1 release artifacts**

Run:

```bash
VERSION=v0.2.1 scripts/build-release.sh --all
```

Expected: `dist/routepeek-v0.2.1-windows-amd64.zip` and `dist/routepeek-v0.2.1-windows-arm64.zip` are regenerated.

- [ ] **Step 4: Windows manual verification**

On Windows, test with proxy and VPN enabled:

```powershell
.\routepeek-web.exe
```

Open the web UI and verify:

- Details page shows `代理: 已开启`.
- Details page shows `VPN: <adapter name>` instead of disconnected.
- Topology includes proxy/VPN nodes when detected.
- Trace page still runs without the previous `tracert timed out after 3s` failure.

---

## Review Checklist

- [ ] Windows system proxy covers WinINet, PAC, WinHTTP, and env vars.
- [ ] VPN detection covers common adapter names without marking Wi-Fi/Ethernet/VM adapters as VPN.
- [ ] Detection helpers are tested with deterministic parser tests.
- [ ] Existing public JSON schema remains unchanged.
- [ ] `go test ./...`, `npm test`, and `VERSION=v0.2.1 scripts/build-release.sh --all` pass.
