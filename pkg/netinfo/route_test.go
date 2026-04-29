package netinfo

import (
	"errors"
	"reflect"
	"testing"

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
