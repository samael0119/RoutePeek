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

func TestParseWindowsWinINetPerSchemeProxy(t *testing.T) {
	got := parseWindowsWinINetProxy(`{"ProxyEnable":1,"ProxyServer":"http=127.0.0.1:7890;https=127.0.0.1:7891","AutoConfigURL":""}`)
	if got == nil || !got.HasProxy || got.HTTPProxy != "127.0.0.1:7890" || got.HTTPSProxy != "127.0.0.1:7891" {
		t.Fatalf("unexpected per-scheme proxy config %#v", got)
	}
}

func TestParseWindowsWinHTTPProxy(t *testing.T) {
	output := "Current WinHTTP proxy settings:\r\n\r\n    Proxy Server(s) : 127.0.0.1:7890\r\n    Bypass List     : localhost;*.local\r\n"
	got := parseWindowsWinHTTPProxy(output)
	if got == nil || !got.HasProxy || got.HTTPProxy != "127.0.0.1:7890" || got.NoProxy != "localhost;*.local" {
		t.Fatalf("unexpected winhttp proxy config %#v", got)
	}
}

func TestParseWindowsWinHTTPPerSchemeProxy(t *testing.T) {
	output := "Current WinHTTP proxy settings:\r\n\r\n    Proxy Server(s) : http=127.0.0.1:7890;https=127.0.0.1:7891\r\n"
	got := parseWindowsWinHTTPProxy(output)
	if got == nil || !got.HasProxy || got.HTTPProxy != "127.0.0.1:7890" || got.HTTPSProxy != "127.0.0.1:7891" {
		t.Fatalf("unexpected winhttp per-scheme proxy config %#v", got)
	}
}

func TestParseWindowsWinHTTPDirectAccess(t *testing.T) {
	output := "Current WinHTTP proxy settings:\r\n\r\n    Direct access (no proxy server).\r\n"
	if got := parseWindowsWinHTTPProxy(output); got != nil && got.HasProxy {
		t.Fatalf("expected no proxy, got %#v", got)
	}
}

func TestDetectProxyUsesAllProxyForHTTPAndHTTPS(t *testing.T) {
	for _, env := range ProxyEnvVars {
		t.Setenv(env, "")
	}
	t.Setenv("ALL_PROXY", "socks5://127.0.0.1:7890")

	got := DetectProxy()
	if got == nil || !got.HasProxy {
		t.Fatalf("expected proxy from ALL_PROXY, got %#v", got)
	}
	if got.HTTPProxy != "socks5://127.0.0.1:7890" || got.HTTPSProxy != "socks5://127.0.0.1:7890" {
		t.Fatalf("expected ALL_PROXY to populate http/https proxies, got %#v", got)
	}
}

func TestDetectProxyUsesNoProxy(t *testing.T) {
	for _, env := range ProxyEnvVars {
		t.Setenv(env, "")
	}
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")
	t.Setenv("NO_PROXY", "localhost,127.0.0.1")

	got := DetectProxy()
	if got == nil || got.NoProxy != "localhost,127.0.0.1" {
		t.Fatalf("expected NO_PROXY to be preserved, got %#v", got)
	}
}
