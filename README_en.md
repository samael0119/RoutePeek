# RoutePeek - Network Topology Visualization and Diagnostic Tool

English | [中文](./README.md)

A cross-platform network configuration visualization tool for beginners, helping users understand complex network environments like multi-interface, VPN, proxy, and virtual machines.

## Core Features

| Feature | Description |
|---------|-------------|
| Multi-Interface Detection | Simultaneously show all interfaces: Ethernet, Wi-Fi, VPN, VM networks, etc. |
| Dynamic SVG Topology | Automatically render layered network paths with animated flow for clear data visualization. |
| i18n Support | Built-in English and Chinese support, auto-detected based on system environment. |
| Plain Language Explanations | "Translated" technical terms into simple language that anyone can understand. |
| Smart Diagnostic Linking | Diagnostic findings pulse on the topology map to highlight problem locations. |
| VPN/Proxy Awareness | Detect global VPN, proxy configs, and their impact on LAN/VM networks. |
| Traceroute | `/api/trace?target=` Scenario-based path tracing to any target. |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI / Web UI                        │
├─────────────────────────────────────────────────────────────┤
│  cmd/cli/main.go          │  cmd/server/main.go            │
│  - Cobra CLI              │  - HTTP API + Static Pages     │
│  - scan/diag/routes        │  - Gorilla/Mux Routing        │
└─────────────┬───────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────┐
│                   internal/discovery                        │
│  - GetNetworkSnapshot()  Aggregate all network info         │
│  - RunDiagnostics()      Execute diagnostic checks          │
│  - PrintNetworkOverview() CLI formatted output              │
└─────────────┬───────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────┐
│                     pkg/netinfo                              │
│  Cross-platform info collection (Pure Go, no cgo)           │
│                                                             │
│  interface.go   - GetNetworkInterfaces()                   │
│  route.go       - GetRoutes() / GetDefaultGateway()         │
│  dns.go        - GetDNSConfig()                             │
│  proxy.go      - DetectProxy()                              │
│  vpn.go        - DetectVPN() / DetectVMNetworks()           │
│  consts.go     - VPNInterfacePrefixes / KnownVMSubnets     │
└─────────────┬───────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────┐
│                      pkg/types                               │
│  NetworkInterface / RouteEntry / DNSConfig / ProxyConfig    │
│  VPNInfo / VMNetwork / NetworkSnapshot / DiagnosisReport     │
└─────────────────────────────────────────────────────────────┘
```

## Technical Solution

### Cross-platform Implementation
- **Linux/macOS/Windows** unified interface, implemented via shell commands.
- Avoids cgo dependencies for easy cross-compilation.
- Interface enumeration: `net.Interfaces()` + platform-specific commands.
- Routing table: `ip route` / `netstat -rn` / `route print`.
- DNS configuration: `/etc/resolv.conf` / `scutil --dns` / `ipconfig /all`.
- VPN detection: Scan for `tun`/`tap`/`wg`/`utun` prefixes + configuration directory scanning.

### Diagnostic Engine
- DNS anomaly detection: Compare against known public DNS lists (Google, Cloudflare, etc.).
- Routing conflict detection: Multiple routes for the same target network.
- VPN LAN impact detection: 0.0.0.0/0 route covering VM subnets (192.168.56.0/24, etc.).
- VM segment connectivity detection: Probe Docker/VMware/VirtualBox/Hyper-V segments.

### Security Mechanisms
- One-click fix (Phase 2): Backup → Confirmation → Auto-rollback.
- Diagnostic findings include specific fix suggestions, not just warnings.

## Quick Start

```bash
# Build
go build -o RoutePeek ./cmd/cli
go build -o RoutePeek-web ./cmd/server

# CLI Usage
./RoutePeek scan          # View full network configuration
./RoutePeek diag          # Run network diagnostics
./RoutePeek routes        # View routing table
./RoutePeek interfaces    # View interface details

# Web Interface
./RoutePeek-web           # Start web server (Default port 8080)
```

## Interface Overview

### 1. Network Panorama (Home)
- **Status Summary**: Natural language summary of current network status (e.g., "✅ You are connecting to the Internet via Ethernet").
- **SVG Layered Topology**: Dynamic rendering of LAN, Gateway, and Internet layers with animated connections.
- **Diagnostic Panel**: Displays reasons and fix suggestions for detected anomalies.

### 2. Detailed Information
For advanced users to view complete technical parameters:
- Detailed interface configs (MAC, MTU, Status).
- System routing table (Destination, Gateway, Metric).
- DNS servers and proxy settings.

### 3. Path Tracing
- Preset scenarios (Test Baidu, Test Google, Test DNS).
- Color-coded hops (Local Network, Carrier Backbone, Target Server) with plain-language explanations.

## Project Structure

```
RoutePeek/
├── cmd/
│   ├── cli/main.go          # CLI Entry (Cobra)
│   └── server/              # Web Server
│       ├── main.go          # Server entry
│       └── static/          # Web UI static assets
├── internal/
│   ├── diagnostic/          # Diagnostic engine
│   ├── discovery/           # Network info aggregation
│   └── i18n/                # Internationalization module
├── pkg/
│   ├── netinfo/             # Cross-platform collection
│   └── types/               # Data structures
├── docs/                    # Documentation and plans
├── go.mod
└── README.md
```

## Dependencies

```
github.com/gorilla/mux v1.8.1    # HTTP Routing
github.com/spf13/cobra v1.8.0     # CLI Framework
```

## Future Plans

- [x] SVG Visualization in Web UI
- [ ] Phase 2: One-click fix (Backup + Confirmation + Rollback)
- [ ] Migrate Web UI to Vue/React (Improve responsiveness and maintainability)
- [ ] Create Makefile for one-click cross-platform builds
- [ ] Support DNS over HTTPS/TLS detection
- [ ] Integrate traceroute geolocation into detailed lists
- [ ] Export diagnostic reports as PDF/Images

## Background

Tools like MTR/WinMTR are for advanced users. Pure CLI output is unfriendly for beginners. This tool's core value is **simultaneously showing all network layers** (Ethernet/Wi-Fi/VPN/Proxy/VM) and explaining them in plain language, especially for troubleshooting "VPN affecting VM connection" issues.
