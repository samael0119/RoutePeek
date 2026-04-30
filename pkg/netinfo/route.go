package netinfo

import (
	"bufio"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/samael0119/RoutePeek/pkg/types"
)

var tcpPortProbe = func(target string, port string) (bool, string) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, port), 1500*time.Millisecond)
	if err != nil {
		return false, ""
	}
	_ = conn.Close()
	return true, fmt.Sprintf("%.1f ms", float64(time.Since(start).Microseconds())/1000)
}

var lookupIP = net.LookupIP

// GetRoutes returns the routing table entries
func GetRoutes() ([]types.RouteEntry, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxRoutes()
	case "darwin":
		return getMacRoutes()
	case "windows":
		return getWindowsRoutes()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// GetDefaultGateway returns the default gateway
func GetDefaultGateway() (string, error) {
	routes, err := GetRoutes()
	if err != nil {
		return "", err
	}

	for _, route := range routes {
		if route.Destination == "0.0.0.0/0" || route.Destination == "::/0" {
			return route.Gateway, nil
		}
	}

	return "", fmt.Errorf("no default gateway found")
}

// GetDefaultInterface returns the interface used for default route
func GetDefaultInterface() (string, error) {
	routes, err := GetRoutes()
	if err != nil {
		return "", err
	}

	for _, route := range routes {
		if route.Destination == "0.0.0.0/0" {
			return route.Interface, nil
		}
	}

	return "", fmt.Errorf("no default route found")
}

// Linux route parsing
func getLinuxRoutes() ([]types.RouteEntry, error) {
	output, err := commandRunner("ip", "route", "show")
	if err != nil {
		return nil, fmt.Errorf("failed to run 'ip route': %w", err)
	}

	var routes []types.RouteEntry
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		entry := types.RouteEntry{
			Table: "main",
		}

		// First field is usually destination
		if parts[0] == "default" {
			entry.Destination = "0.0.0.0/0"
		} else if strings.Contains(parts[0], "/") {
			entry.Destination = parts[0]
		}

		// Parse rest of the line
		for i := 1; i < len(parts); i++ {
			switch parts[i] {
			case "via":
				if i+1 < len(parts) {
					entry.Gateway = parts[i+1]
				}
			case "dev":
				if i+1 < len(parts) {
					entry.Interface = parts[i+1]
				}
			case "metric":
				if i+1 < len(parts) {
					if m, err := strconv.Atoi(parts[i+1]); err == nil {
						entry.Metric = m
					}
				}
			}
		}

		if entry.Destination != "" && entry.Interface != "" {
			routes = append(routes, entry)
		}
	}

	return routes, nil
}

// macOS route parsing
func getMacRoutes() ([]types.RouteEntry, error) {
	output, err := commandRunner("netstat", "-rn")
	if err != nil {
		return nil, fmt.Errorf("failed to run 'netstat -rn': %w", err)
	}

	var routes []types.RouteEntry
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	inIPv4Section := false

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header lines
		if strings.HasPrefix(line, "Internet") || strings.HasPrefix(line, "Internet6") {
			inIPv4Section = strings.HasPrefix(line, "Internet") && !strings.HasPrefix(line, "Internet6")
			continue
		}

		if !inIPv4Section {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		entry := types.RouteEntry{
			Destination: parts[0],
			Gateway:     parts[1],
			Interface:   parts[3],
			Table:       "main",
		}

		if len(parts) >= 7 {
			if m, err := strconv.Atoi(parts[6]); err == nil {
				entry.Metric = m
			}
		}

		if entry.Destination != "" {
			routes = append(routes, entry)
		}
	}

	return routes, nil
}

// Windows route parsing
func getWindowsRoutes() ([]types.RouteEntry, error) {
	output, err := commandRunner("route", "print", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to run 'route print': %w", err)
	}

	var routes []types.RouteEntry
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	capture := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Start capturing when we see IPv4 Route table header
		if strings.Contains(line, "IPv4 Route Table") || strings.Contains(line, "IPv4") {
			capture = true
			continue
		}

		if strings.Contains(line, "IPv6 Route Table") || strings.Contains(line, "IPv6") {
			capture = false
			continue
		}

		if !capture {
			continue
		}

		// Parse route entries (typical format: Network Destination  Netmask  Gateway  Interface  Metric)
		parts := strings.Fields(line)
		if len(parts) >= 5 {
			entry := types.RouteEntry{
				Destination: parts[0],
				Gateway:     parts[2],
				Interface:   parts[3],
				Table:       "main",
			}

			if len(parts) >= 6 {
				if m, err := strconv.Atoi(parts[5]); err == nil {
					entry.Metric = m
				}
			}

			if entry.Destination != "" && entry.Destination != "Gateway" {
				routes = append(routes, entry)
			}
		}
	}

	return routes, nil
}

// GetInterfaceRoutes returns routes associated with a specific interface
func GetInterfaceRoutes(ifaceName string) ([]types.RouteEntry, error) {
	routes, err := GetRoutes()
	if err != nil {
		return nil, err
	}

	var ifaceRoutes []types.RouteEntry
	for _, route := range routes {
		if route.Interface == ifaceName {
			ifaceRoutes = append(ifaceRoutes, route)
		}
	}

	return ifaceRoutes, nil
}

// IsVPNGlobal checks if VPN routes all traffic (0.0.0.0/0 through VPN)
func IsVPNGlobal(vpnInterface string, routes []types.RouteEntry) bool {
	for _, route := range routes {
		if route.Interface == vpnInterface {
			if route.Destination == "0.0.0.0/0" {
				return true
			}
		}
	}
	return false
}

// GetVMSubnetRoutes returns routes that go to VM networks
func GetVMSubnetRoutes(routes []types.RouteEntry) []types.RouteEntry {
	var vmRoutes []types.RouteEntry

	for _, route := range routes {
		// Check if destination is a known VM subnet
		_, dstNet, _ := net.ParseCIDR(route.Destination)
		if dstNet == nil {
			continue
		}

		for _, vmSubnet := range append(CommonVMwareSubnets, append(CommonVirtualBoxSubnets, CommonDockerSubnets...)...) {
			_, vmNet, _ := net.ParseCIDR(vmSubnet)
			if vmNet != nil && (cidrContainsNetwork(vmNet, dstNet) || cidrContainsNetwork(dstNet, vmNet)) {
				vmRoutes = append(vmRoutes, route)
				break
			}
		}
	}

	return vmRoutes
}

// TraceRoute performs a traceroute to the target and returns hop information
func TraceRoute(target string) ([]types.TraceHop, error) {
	switch runtime.GOOS {
	case "linux":
		return traceRouteLinux(target)
	case "darwin":
		return traceRouteMac(target)
	case "windows":
		return traceRouteWindows(target)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func traceRouteLinux(target string) ([]types.TraceHop, error) {
	output, tracepathErr := commandRunner("tracepath", "-m", "12", target)
	if hops, ok := responsiveTraceOutput(output); ok {
		return annotateTraceMode(hops), nil
	}

	output, tracerouteErr := commandRunner("traceroute", "-n", "-m", "12", "-q", "3", "-w", "1", target)
	if hops, ok := responsiveTraceOutput(output); ok {
		return annotateTraceMode(hops), nil
	}

	output, tcpTracerouteErr := commandRunner("traceroute", "-T", "-p", "443", "-n", "-m", "12", "-q", "3", "-w", "1", target)
	if hops, ok := responsiveTraceOutput(output); ok {
		return annotateTraceMode(hops), nil
	}

	if hops, err := routeProbeFallback(target); err == nil && len(hops) > 0 {
		return hops, nil
	}

	return nil, fmt.Errorf("tracepath failed: %v; traceroute failed: %v; tcp traceroute failed: %w", tracepathErr, tracerouteErr, tcpTracerouteErr)
}

func responsiveTraceOutput(output []byte) ([]types.TraceHop, bool) {
	if len(output) == 0 {
		return nil, false
	}
	hops, err := parseTracerouteOutput(string(output))
	if err != nil || !hasResponsiveHop(hops) {
		return nil, false
	}
	return hops, true
}

func hasResponsiveHop(hops []types.TraceHop) bool {
	for _, hop := range hops {
		if strings.TrimSpace(hop.Address) != "" && strings.TrimSpace(hop.Address) != "*" {
			return true
		}
		if strings.TrimSpace(hop.RTT1) != "" && strings.TrimSpace(hop.RTT1) != "*" {
			return true
		}
		if strings.TrimSpace(hop.RTT2) != "" && strings.TrimSpace(hop.RTT2) != "*" {
			return true
		}
		if strings.TrimSpace(hop.RTT3) != "" && strings.TrimSpace(hop.RTT3) != "*" {
			return true
		}
	}
	return false
}

func annotateTraceMode(hops []types.TraceHop) []types.TraceHop {
	if !isProxyFakeIPTrace(hops) {
		return hops
	}
	for i := range hops {
		hops[i].Mode = "proxy_fake_ip"
	}
	return hops
}

func isProxyFakeIPTrace(hops []types.TraceHop) bool {
	last := lastResponsiveAddress(hops)
	if last == "" {
		return false
	}
	ip := net.ParseIP(last)
	if ip == nil {
		return false
	}
	_, fakeIPNet, err := net.ParseCIDR("198.18.0.0/15")
	return err == nil && fakeIPNet.Contains(ip)
}

func lastResponsiveAddress(hops []types.TraceHop) string {
	for i := len(hops) - 1; i >= 0; i-- {
		address := strings.TrimSpace(hops[i].Address)
		if address != "" && address != "*" {
			return address
		}
	}
	return ""
}

func routeProbeFallback(target string) ([]types.TraceHop, error) {
	routeTarget, err := routeLookupTarget(target)
	routeLookupAvailable := err == nil
	if !routeLookupAvailable {
		routeTarget = target
	}

	hops := []types.TraceHop{}
	if routeLookupAvailable {
		output, err := commandRunner("ip", "route", "get", routeTarget)
		if err == nil {
			via, dev, src := parseIPRouteGet(string(output))
			if via != "" {
				hop := types.TraceHop{Hop: 1, Address: via, Mode: "route_probe"}
				if dev != "" {
					hop.Hostname = via + " via " + dev
				} else {
					hop.Hostname = via
				}
				hops = append(hops, hop)
			} else if src != "" {
				hop := types.TraceHop{Hop: 1, Address: src, Mode: "route_probe"}
				if dev != "" {
					hop.Hostname = src + " via " + dev
				} else {
					hop.Hostname = src
				}
				hops = append(hops, hop)
			}
		}
	}

	for _, port := range []string{"443", "80", "53"} {
		if rtts := tcpProbeSamples(target, port, 3); len(rtts) > 0 {
			hops = append(hops, types.TraceHop{
				Hop:      len(hops) + 1,
				Address:  routeTarget,
				Hostname: routeTarget + " tcp/" + port + " reachable",
				RTT1:     rttAt(rtts, 0),
				RTT2:     rttAt(rtts, 1),
				RTT3:     rttAt(rtts, 2),
				Mode:     "route_probe",
			})
			return hops, nil
		}
	}

	if len(hops) > 0 {
		hops = append(hops, types.TraceHop{
			Hop:      len(hops) + 1,
			Address:  "*",
			Hostname: "target tcp unreachable",
			RTT1:     "*",
			Mode:     "route_probe",
		})
		return hops, nil
	}

	return []types.TraceHop{{
		Hop:      1,
		Address:  routeTarget,
		Hostname: routeTarget + " tcp unreachable",
		RTT1:     "*",
		Mode:     "route_probe",
	}}, nil
}

func tcpProbeSamples(target string, port string, count int) []string {
	rtts := []string{}
	for i := 0; i < count; i++ {
		ok, rtt := tcpPortProbe(target, port)
		if ok && rtt != "" {
			rtts = append(rtts, rtt)
		}
	}
	return rtts
}

func rttAt(rtts []string, index int) string {
	if index < len(rtts) {
		return rtts[index]
	}
	return ""
}

func routeLookupTarget(target string) (string, error) {
	if ip := net.ParseIP(target); ip != nil {
		return target, nil
	}
	ips, err := lookupIP(target)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}
	if len(ips) > 0 {
		return ips[0].String(), nil
	}
	return "", fmt.Errorf("no IP address found for %s", target)
}

func parseIPRouteGet(output string) (via string, dev string, src string) {
	fields := strings.Fields(output)
	for i := 0; i+1 < len(fields); i++ {
		switch fields[i] {
		case "via":
			via = fields[i+1]
		case "dev":
			dev = fields[i+1]
		case "src":
			src = fields[i+1]
		}
	}
	return via, dev, src
}

func traceRouteMac(target string) ([]types.TraceHop, error) {
	// Use traceroute on macOS
	output, err := commandRunner("traceroute", "-n", "-m", "12", "-q", "3", "-w", "1", target)
	if err != nil {
		return nil, fmt.Errorf("traceroute failed: %w", err)
	}

	hops, err := parseTracerouteOutput(string(output))
	if err != nil {
		return nil, err
	}
	return annotateTraceMode(hops), nil
}

func traceRouteWindows(target string) ([]types.TraceHop, error) {
	// Use tracert on Windows
	output, err := commandRunner("tracert", "-d", "-h", "12", "-w", "1000", target)
	if err != nil {
		return nil, fmt.Errorf("tracert failed: %w", err)
	}

	hops, err := parseWindowsTraceOutput(string(output))
	if err != nil {
		return nil, err
	}
	return annotateTraceMode(hops), nil
}

func parseTracerouteOutput(output string) ([]types.TraceHop, error) {
	var hops []types.TraceHop
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		// Skip header lines
		if strings.HasPrefix(line, "traceroute") || strings.HasPrefix(line, "tracepath") {
			continue
		}

		// Parse lines like: " 1  192.168.1.1  1.234 ms  1.456 ms  1.789 ms"
		// or: " 1  192.168.1.1  1.234 ms"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		hop := types.TraceHop{}

		// First field should be hop number. tracepath may suffix it with ":" or "?:".
		hopNumStr := strings.Trim(fields[0], " *?:")
		if n, err := strconv.Atoi(hopNumStr); err == nil {
			hop.Hop = n
		} else {
			continue
		}
		if len(fields) > 1 && strings.Contains(fields[1], "LOCALHOST") {
			continue
		}

		// Find IP address (usually second field, may be * or hostname)
		for i := 1; i < len(fields); i++ {
			f := fields[i]
			if f == "*" {
				hop.Address = "*"
				break
			}
			// Check if it looks like an IP address
			if net.ParseIP(f) != nil {
				hop.Address = f
				break
			}
			// Could be hostname followed by IP in parens
			if strings.Contains(f, "(") {
				// hostname(ip)
				if idx := strings.Index(f, "("); idx > 0 {
					hop.Hostname = f[:idx]
				}
				ip := strings.Trim(strings.Trim(f, "()"), "*")
				if net.ParseIP(ip) != nil {
					hop.Address = ip
				}
				break
			}
		}

		// Parse RTT values - combine adjacent number + unit fields
		// e.g. ["2.677", "ms"] -> "2.677 ms"
		for i := 1; i < len(fields); i++ {
			f := strings.TrimRight(fields[i], "\r")
			if strings.HasSuffix(f, "ms") || strings.HasSuffix(f, "s") {
				unit := "ms"
				raw := strings.TrimSuffix(f, "ms")
				if strings.HasSuffix(f, "s") && !strings.HasSuffix(f, "ms") {
					unit = "s"
					raw = strings.TrimSuffix(f, "s")
				}
				if _, err := strconv.ParseFloat(raw, 64); err == nil {
					val := raw + " " + unit
					switch {
					case hop.RTT1 == "":
						hop.RTT1 = val
					case hop.RTT2 == "":
						hop.RTT2 = val
					case hop.RTT3 == "":
						hop.RTT3 = val
					}
				}
				continue
			}
			if i+1 >= len(fields) {
				continue
			}
			unit := strings.TrimRight(fields[i+1], "\r")
			if _, err := strconv.ParseFloat(f, 64); err == nil {
				if unit == "ms" || unit == "s" {
					val := f + " " + unit
					switch {
					case hop.RTT1 == "":
						hop.RTT1 = val
					case hop.RTT2 == "":
						hop.RTT2 = val
					case hop.RTT3 == "":
						hop.RTT3 = val
					}
				}
			}
		}

		if hop.Hop > 0 {
			hops = append(hops, hop)
		}
	}

	return hops, nil
}

func cidrContainsNetwork(parent, child *net.IPNet) bool {
	if parent == nil || child == nil {
		return false
	}
	if !parent.Contains(child.IP) {
		return false
	}
	onesParent, bitsParent := parent.Mask.Size()
	onesChild, bitsChild := child.Mask.Size()
	return bitsParent == bitsChild && onesParent <= onesChild
}

func parseWindowsTraceOutput(output string) ([]types.TraceHop, error) {
	var hops []types.TraceHop
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header lines
		if strings.HasPrefix(strings.TrimSpace(line), "Tracing") {
			continue
		}

		// Parse lines like: "  1     1 ms     1 ms     1 ms  192.168.1.1"
		// or: "  1        *        *        *     请求超时"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		hop := types.TraceHop{}

		// First field is hop number
		if n, err := strconv.Atoi(strings.Trim(fields[0], " ")); err == nil {
			hop.Hop = n
		} else {
			continue
		}

		// Find IP address - usually last field
		for i := len(fields) - 1; i >= 0; i-- {
			f := fields[i]
			if f == "*" || strings.Contains(f, "超时") || strings.Contains(f, "timed") {
				hop.Address = "*"
				continue
			}
			if net.ParseIP(f) != nil {
				hop.Address = f
				break
			}
		}

		// Parse RTT values - fields between hop number and IP
		for i := 1; i < len(fields); i++ {
			f := fields[i]
			if strings.HasSuffix(f, "ms") || strings.HasSuffix(f, "s") {
				val := strings.Trim(strings.Trim(f, "ms"), "s")
				switch {
				case hop.RTT1 == "":
					hop.RTT1 = val + " ms"
				case hop.RTT2 == "":
					hop.RTT2 = val + " ms"
				case hop.RTT3 == "":
					hop.RTT3 = val + " ms"
				}
			}
		}

		if hop.Hop > 0 {
			hops = append(hops, hop)
		}
	}

	return hops, nil
}
