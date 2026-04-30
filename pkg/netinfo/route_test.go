package netinfo

import (
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/samael0119/RoutePeek/pkg/types"
)

func TestGetVMSubnetRoutesMatchesContainedCIDR(t *testing.T) {
	routes := []types.RouteEntry{
		{Destination: "172.17.4.0/24", Interface: "docker0"},
		{Destination: "172.16.0.0/12", Interface: "hyperv"},
		{Destination: "192.0.2.0/24", Interface: "eth0"},
	}

	got := GetVMSubnetRoutes(routes)
	if len(got) != 2 {
		t.Fatalf("expected two VM routes, got %d: %#v", len(got), got)
	}
	if got[0].Destination != "172.17.4.0/24" {
		t.Fatalf("expected contained Docker route, got %q", got[0].Destination)
	}
	if got[1].Destination != "172.16.0.0/12" {
		t.Fatalf("expected broader Hyper-V route, got %q", got[1].Destination)
	}
}

func TestParseTracerouteOutputSupportsTracepath(t *testing.T) {
	output := ` 1?: [LOCALHOST]                      pmtu 1500
 1:  192.168.1.1                                           1.431ms
 2:  10.0.0.1                                               8.042ms
`

	got, err := parseTracerouteOutput(output)
	if err != nil {
		t.Fatalf("parseTracerouteOutput returned error: %v", err)
	}

	want := []types.TraceHop{
		{Hop: 1, Address: "192.168.1.1", RTT1: "1.431 ms"},
		{Hop: 2, Address: "10.0.0.1", RTT1: "8.042 ms"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tracepath parse mismatch\nwant: %#v\n got: %#v", want, got)
	}
}

func TestTraceRouteLinuxFallsBackToOrdinaryTraceroute(t *testing.T) {
	orig := commandRunner
	t.Cleanup(func() { commandRunner = orig })

	var calls []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, name)
		switch name {
		case "tracepath":
			return nil, errors.New("not installed")
		case "traceroute":
			return []byte(" 1  192.168.1.1  1.2 ms\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}

	hops, err := traceRouteLinux("example.com")
	if err != nil {
		t.Fatalf("traceRouteLinux returned error: %v", err)
	}
	if len(hops) != 1 || hops[0].Address != "192.168.1.1" {
		t.Fatalf("unexpected hops: %#v", hops)
	}

	wantCalls := []string{"tracepath", "traceroute"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("command order mismatch: want %v got %v", wantCalls, calls)
	}
}

func TestTraceRouteLinuxUsesBoundedProbeSettings(t *testing.T) {
	orig := commandRunner
	t.Cleanup(func() { commandRunner = orig })

	var tracepathArgs []string
	var tracerouteArgs []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		switch name {
		case "tracepath":
			tracepathArgs = append([]string(nil), args...)
			return nil, errors.New("not installed")
		case "traceroute":
			tracerouteArgs = append([]string(nil), args...)
			return []byte(" 1  192.168.1.1  1.2 ms\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}

	if _, err := traceRouteLinux("example.com"); err != nil {
		t.Fatalf("traceRouteLinux returned error: %v", err)
	}

	if !reflect.DeepEqual(tracepathArgs, []string{"-m", "12", "example.com"}) {
		t.Fatalf("unexpected tracepath args: %v", tracepathArgs)
	}
	wantTracerouteArgs := []string{"-n", "-m", "12", "-q", "3", "-w", "1", "example.com"}
	if !reflect.DeepEqual(tracerouteArgs, wantTracerouteArgs) {
		t.Fatalf("unexpected traceroute args: want %v got %v", wantTracerouteArgs, tracerouteArgs)
	}
}

func TestTraceRouteLinuxMarksProxyFakeIPResults(t *testing.T) {
	orig := commandRunner
	t.Cleanup(func() { commandRunner = orig })

	commandRunner = func(name string, args ...string) ([]byte, error) {
		switch name {
		case "tracepath":
			return nil, errors.New("not installed")
		case "traceroute":
			return []byte(" 1  198.18.0.29  0.765 ms  0.812 ms  0.779 ms\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}

	hops, err := traceRouteLinux("baidu.com")
	if err != nil {
		t.Fatalf("traceRouteLinux returned error: %v", err)
	}

	want := []types.TraceHop{
		{Hop: 1, Address: "198.18.0.29", RTT1: "0.765 ms", RTT2: "0.812 ms", RTT3: "0.779 ms", Mode: "proxy_fake_ip"},
	}
	if !reflect.DeepEqual(hops, want) {
		t.Fatalf("proxy fake-ip hops mismatch\nwant: %#v\n got: %#v", want, hops)
	}
}

func TestTraceRouteLinuxUsesPartialTracerouteOutputOnTimeout(t *testing.T) {
	orig := commandRunner
	t.Cleanup(func() { commandRunner = orig })

	commandRunner = func(name string, args ...string) ([]byte, error) {
		switch name {
		case "tracepath":
			return nil, errors.New("tracepath timed out")
		case "traceroute":
			return []byte(" 1  192.168.1.1  1.0 ms  1.1 ms  1.2 ms\n 2  203.0.113.1  20.1 ms  20.2 ms  20.3 ms\n"), errors.New("traceroute timed out")
		default:
			return nil, errors.New("unexpected command")
		}
	}

	hops, err := traceRouteLinux("baidu.com")
	if err != nil {
		t.Fatalf("traceRouteLinux returned error: %v", err)
	}

	want := []types.TraceHop{
		{Hop: 1, Address: "192.168.1.1", RTT1: "1.0 ms", RTT2: "1.1 ms", RTT3: "1.2 ms"},
		{Hop: 2, Address: "203.0.113.1", RTT1: "20.1 ms", RTT2: "20.2 ms", RTT3: "20.3 ms"},
	}
	if !reflect.DeepEqual(hops, want) {
		t.Fatalf("partial timeout hops mismatch\nwant: %#v\n got: %#v", want, hops)
	}
}

func TestTraceRouteLinuxFallsBackToRouteProbeWhenCommandsOnlyTimeout(t *testing.T) {
	origRunner := commandRunner
	origTCPProbe := tcpPortProbe
	t.Cleanup(func() {
		commandRunner = origRunner
		tcpPortProbe = origTCPProbe
	})

	var calls []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, name+" "+strings.Join(args, " "))
		switch name {
		case "tracepath":
			return []byte(" 1:  no reply\n 2:  no reply\n"), nil
		case "traceroute":
			if len(args) > 0 && args[0] == "-T" {
				return nil, errors.New("operation not permitted")
			}
			return []byte(" 1  *\n 2  *\n"), nil
		case "ip":
			return []byte("8.8.8.8 via 198.18.0.2 dev Mihomo src 198.18.0.1 uid 1000\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}
	probes := []string{"12.3 ms", "13.4 ms", "11.9 ms"}
	tcpPortProbe = func(target string, port string) (bool, string) {
		if target == "8.8.8.8" && port == "443" {
			rtt := probes[0]
			probes = probes[1:]
			return true, rtt
		}
		return false, ""
	}

	hops, err := traceRouteLinux("8.8.8.8")
	if err != nil {
		t.Fatalf("traceRouteLinux returned error: %v", err)
	}

	want := []types.TraceHop{
		{Hop: 1, Address: "198.18.0.2", Hostname: "198.18.0.2 via Mihomo", Mode: "route_probe"},
		{Hop: 2, Address: "8.8.8.8", Hostname: "8.8.8.8 tcp/443 reachable", RTT1: "12.3 ms", RTT2: "13.4 ms", RTT3: "11.9 ms", Mode: "route_probe"},
	}
	if !reflect.DeepEqual(hops, want) {
		t.Fatalf("fallback hops mismatch\nwant: %#v\n got: %#v", want, hops)
	}
	if len(calls) != 4 || !strings.HasPrefix(calls[3], "ip route get 8.8.8.8") {
		t.Fatalf("expected route lookup after trace command timeouts, got calls %v", calls)
	}
}

func TestTraceRouteLinuxFallsBackToTCPProbeWhenHostnameLookupFails(t *testing.T) {
	origRunner := commandRunner
	origTCPProbe := tcpPortProbe
	origLookupIP := lookupIP
	t.Cleanup(func() {
		commandRunner = origRunner
		tcpPortProbe = origTCPProbe
		lookupIP = origLookupIP
	})

	commandRunner = func(name string, args ...string) ([]byte, error) {
		switch name {
		case "tracepath", "traceroute":
			return nil, errors.New(name + " timed out")
		default:
			return nil, errors.New("unexpected command")
		}
	}
	lookupIP = func(host string) ([]net.IP, error) {
		return nil, errors.New("temporary failure in name resolution")
	}
	probes := []string{"31.2 ms", "30.8 ms", "32.1 ms"}
	tcpPortProbe = func(target string, port string) (bool, string) {
		if target == "baidu.com" && port == "443" {
			rtt := probes[0]
			probes = probes[1:]
			return true, rtt
		}
		return false, ""
	}

	hops, err := traceRouteLinux("baidu.com")
	if err != nil {
		t.Fatalf("traceRouteLinux returned error: %v", err)
	}

	want := []types.TraceHop{
		{Hop: 1, Address: "baidu.com", Hostname: "baidu.com tcp/443 reachable", RTT1: "31.2 ms", RTT2: "30.8 ms", RTT3: "32.1 ms", Mode: "route_probe"},
	}
	if !reflect.DeepEqual(hops, want) {
		t.Fatalf("hostname fallback hops mismatch\nwant: %#v\n got: %#v", want, hops)
	}
}

func TestParseWindowsRoutePrintOutputSkipsLocalizedHeaders(t *testing.T) {
	output := `===========================================================================
IPv4 Route Table
===========================================================================
Active Routes:
Network Destination        Netmask          Gateway       Interface  Metric
����Ŀ��                  ����            ����          �ӿ�        Ծ����
          0.0.0.0          0.0.0.0    192.168.1.1  192.168.1.23     25
        127.0.0.0        255.0.0.0         On-link      127.0.0.1    331
===========================================================================
`

	got := parseWindowsRoutePrintOutput(output)

	want := []types.RouteEntry{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Interface: "192.168.1.23", Metric: 25, Table: "main"},
		{Destination: "127.0.0.0/8", Gateway: "On-link", Interface: "127.0.0.1", Metric: 331, Table: "main"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("windows routes mismatch\nwant: %#v\n got: %#v", want, got)
	}
}

func TestParseWindowsTraceOutputCombinesSplitRTTFields(t *testing.T) {
	output := `Tracing route to 8.8.8.8 over a maximum of 12 hops

  1    <1 ms    <1 ms     1 ms  192.168.1.1
  2     *        *        *     Request timed out.
`

	got, err := parseWindowsTraceOutput(output)
	if err != nil {
		t.Fatalf("parseWindowsTraceOutput returned error: %v", err)
	}

	want := []types.TraceHop{
		{Hop: 1, Address: "192.168.1.1", RTT1: "<1 ms", RTT2: "<1 ms", RTT3: "1 ms"},
		{Hop: 2, Address: "*"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("windows trace mismatch\nwant: %#v\n got: %#v", want, got)
	}
}

func TestTraceRouteWindowsUsesLongerTimeoutAndPartialOutput(t *testing.T) {
	orig := commandRunnerWithTimeout
	t.Cleanup(func() { commandRunnerWithTimeout = orig })

	var gotTimeout time.Duration
	commandRunnerWithTimeout = func(timeout time.Duration, name string, args ...string) ([]byte, error) {
		gotTimeout = timeout
		if name != "tracert" {
			t.Fatalf("unexpected command %q", name)
		}
		return []byte("  1     1 ms     1 ms     1 ms  192.168.1.1\n"), errors.New("tracert timed out")
	}

	hops, err := traceRouteWindows("8.8.8.8")
	if err != nil {
		t.Fatalf("traceRouteWindows returned error: %v", err)
	}
	if gotTimeout <= commandTimeout {
		t.Fatalf("expected Windows tracert timeout to exceed default command timeout, got %s", gotTimeout)
	}
	want := []types.TraceHop{{Hop: 1, Address: "192.168.1.1", RTT1: "1 ms", RTT2: "1 ms", RTT3: "1 ms"}}
	if !reflect.DeepEqual(hops, want) {
		t.Fatalf("partial Windows trace mismatch\nwant: %#v\n got: %#v", want, hops)
	}
}
