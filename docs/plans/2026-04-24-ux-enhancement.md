# RoutePeek UX Enhancement Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Transform RoutePeek from a technical network tool into a beginner-friendly network topology visualizer with SVG diagrams, plain-language explanations, diagnostic integration, and improved traceroute experience.

**Architecture:**渐进增强现有单文件 `index.html`，用 inline SVG 替换 div 列表拓扑、新增通俗解释映射表、诊断结果与拓扑联动、Traceroute 场景化改造。后端最小改动：扩充 VPN 前缀 + 后备检测逻辑 + 新增诊断项。

**Tech Stack:** Go (backend, minimal changes) · Vanilla HTML/CSS/JS with inline SVG (frontend) · No new dependencies

**Design doc:** `docs/plans/2026-04-24-ux-enhancement-design.md`

---

### Task 1: Expand VPN Detection (Backend)

**Files:**
- Modify: `pkg/netinfo/consts.go`
- Modify: `pkg/netinfo/vpn.go`

**Step 1: Update VPN interface prefixes in consts.go**

Add new VPN prefixes to `VPNInterfacePrefixes`:

```go
// In consts.go, update VPNInterfacePrefixes
VPNInterfacePrefixes = []string{
    "tun", "tap", "ppp", "vpnc", "vpn", "wg", "utun", "ipsec",
    "cnem", "cnem_vnic",
    // New additions
    "tailscale", "nordlynx", "warp", "mullvad", "proton", "wireguard",
}
```

**Step 2: Add fallback VPN detection in vpn.go**

After the prefix-based scan, add a route-based fallback: if an interface has a `0.0.0.0/0` route but is not the default gateway interface, mark it as suspected VPN.

```go
// Add to DetectVPN(), after prefix scan loop, before setting "disconnected":
// Fallback: check routes for non-default-gateway 0.0.0.0/0 routes
routes, _ := GetRoutes()
defaultGw, _ := GetDefaultGateway()
for _, route := range routes {
    if route.Destination == "default" || route.Destination == "0.0.0.0/0" {
        if route.Interface != "" && route.Gateway != defaultGw {
            // This interface routes all traffic but isn't the default gateway - likely VPN
            vpn.Interface = route.Interface
            vpn.Status = "connected"
            vpn.Name = route.Interface + " (detected)"
            // Try to get client IP
            if iface, err := net.InterfaceByName(route.Interface); err == nil {
                addrs, _ := iface.Addrs()
                for _, addr := range addrs {
                    if v, ok := addr.(*net.IPNet); ok && v.IP.To4() != nil {
                        vpn.ClientIP = v.IP.String()
                        vpn.Protocol = guessVPNProtocol(v.IP)
                    }
                }
            }
            return vpn
        }
    }
}
```

**Step 3: Build and verify**

Run: `cd /home/hyh/Projects/routepeek && go build ./...`
Expected: Successful build with no errors

**Step 4: Commit**

```bash
git add pkg/netinfo/consts.go pkg/netinfo/vpn.go
git commit -m "feat: expand VPN detection with new prefixes and route-based fallback"
```

---

### Task 2: Add New Diagnostic Checks (Backend)

**Files:**
- Modify: `internal/diagnostic/diagnose.go`

**Step 1: Add multi-NIC conflict detection**

```go
func diagnoseMultiNICConflict(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
    // Count active non-loopback, non-vm interfaces
    activeIfaces := []string{}
    for _, iface := range snapshot.Interfaces {
        if iface.IsUp && !iface.IsLoopback && iface.IP4 != "" &&
           iface.Type != "vm" && iface.Type != "docker" && iface.Type != "loopback" {
            activeIfaces = append(activeIfaces, iface.Name)
        }
    }
    if len(activeIfaces) > 1 {
        report.Findings = append(report.Findings, types.DiagnosisResult{
            Severity:   "info",
            Code:       "MULTI_NIC_ACTIVE",
            Title:      "多张网卡同时活跃",
            Message:    fmt.Sprintf("检测到 %d 张网卡同时活跃：%s。\n多网卡可能导致路由选择不确定，部分流量走错网卡。", len(activeIfaces), strings.Join(activeIfaces, "、")),
            Suggestion: "如果不需要多网卡，可以禁用不使用的网卡，避免路由冲突。",
        })
    }
}
```

**Step 2: Add no-public-IP detection**

```go
func diagnoseNoPublicIP(snapshot *types.NetworkSnapshot, report *types.DiagnosisReport) {
    if snapshot.PublicIP == "" {
        report.Findings = append(report.Findings, types.DiagnosisResult{
            Severity:   "warning",
            Code:       "NO_PUBLIC_IP",
            Title:      "无法获取公网 IP",
            Message:    "无法从外部服务获取你的公网 IP 地址。\n可能原因：网络未连接、防火墙阻止、或处于严格内网环境。",
            Suggestion: "检查网络连接是否正常，尝试在浏览器中访问任意网站确认。",
        })
    }
}
```

**Step 3: Wire new checks into RunDiagnostics**

Add calls to the new functions in `RunDiagnostics()`:

```go
diagnoseMultiNICConflict(snapshot, report)
diagnoseNoPublicIP(snapshot, report)
```

**Step 4: Build and verify**

Run: `cd /home/hyh/Projects/routepeek && go build ./...`
Expected: Successful build

**Step 5: Commit**

```bash
git add internal/diagnostic/diagnose.go
git commit -m "feat: add multi-NIC conflict and no-public-IP diagnostics"
```

---

### Task 3: Frontend — CSS Foundation for New Layout

**Files:**
- Modify: `cmd/server/static/index.html` (CSS section only)

**Step 1: Add new CSS variables and styles for SVG topology**

Add to the `<style>` section:
- SVG topology container styles (`.svg-topology`)
- Node styles for SVG-rendered nodes (`.topo-svg-node`, layer-specific)
- Flowing connection animation (`@keyframes flowDash`)
- Pulse animation for diagnostic nodes (`@keyframes diagPulse`)
- Status summary bar styles (`.status-summary`)
- Detail panel styles (`.detail-panel`, expandable)
- Traceroute preset buttons (`.trace-presets`)
- RTT color bar (`.rtt-bar`)
- Tab structure: 3 tabs instead of 5

**Step 2: Update HTML structure**

- Change nav tabs from 5 to 3: 网络全景 / 详细信息 / 路径追踪
- Add status summary bar below header
- Add SVG container `<svg id="topoSvg">` inside 网络全景 section
- Move interfaces + routes + DNS into 详细信息 tab
- Add trace preset buttons to 路径追踪 tab
- Add detail panel div for node click expansion

**Step 3: Verify page loads**

Run: `cd /home/hyh/Projects/routepeek && go build -o routepeek-web ./cmd/server && ./routepeek-web`
Expected: Page loads at http://localhost:8080 without errors (may look broken visually at this stage)

**Step 4: Commit**

```bash
git add cmd/server/static/index.html
git commit -m "feat: restructure HTML/CSS for 3-tab layout and SVG topology"
```

---

### Task 4: Frontend — Explanation System

**Files:**
- Modify: `cmd/server/static/index.html` (JS section)

**Step 1: Add explanation mapping object**

At the top of the `<script>` section, add the full `explanations` object covering:
- `interfaceType`: ethernet, wifi, vpn, vm, docker, loopback, vmware, unknown
- `networkConcepts`: gateway, dns, publicIP, vpn, proxy, internet
- `dnsExplanation(ip)`: returns explanation based on whether it's public/private/known
- `generateStatusSummary(snapshot, diagnosis)`: returns natural language one-liner

**Step 2: Add status summary renderer**

Implement `generateStatusSummary()` that analyzes the snapshot and builds a natural language sentence describing current network state:
- Which interface is primary (has default route)
- Whether VPN is active
- Whether proxy is configured
- Any diagnostic warnings

**Step 3: Verify explanation rendering**

Run the server, check that the status bar shows a natural language summary.
Expected: Status bar like "✅ 你正通过有线网卡连接路由器上网"

**Step 4: Commit**

```bash
git add cmd/server/static/index.html
git commit -m "feat: add explanation system and natural language status bar"
```

---

### Task 5: Frontend — SVG Topology Renderer

**Files:**
- Modify: `cmd/server/static/index.html` (JS section)

This is the largest task. Implement `renderSVGTopology(snapshot, diagnosis)`:

**Step 1: Layout calculation**

Write helper functions:
- `calcNodePositions(snapshot)`: returns `{layer1: [...], layer2: [...], layer3: [...]}` with x,y coordinates
- Each layer is a horizontal row, nodes are evenly distributed within the row
- SVG viewBox dynamically sized based on node count

**Step 2: SVG node rendering**

Write `createSVGNode(type, label, ip, x, y, color, hasDiagIssue)`:
- Renders a rounded rect with icon, label, and IP
- If `hasDiagIssue`, adds pulsing border animation
- Returns SVG group element string

**Step 3: SVG connection rendering**

Write `createSVGConnection(fromX, fromY, toX, toY, color)`:
- Renders bezier curve `<path>` between nodes
- Adds `stroke-dasharray` flow animation
- Returns SVG path element string

**Step 4: Assemble full topology**

Write `renderSVGTopology(snapshot, diagnosis)`:
- Layer 1: computer node center + all interface nodes spread horizontally
- Layer 2: gateway, VPN, proxy, DNS nodes
- Layer 3: public IP + internet cloud
- Connect layers with animated paths
- Mark diagnostic issue nodes with pulse animation

**Step 5: Node interaction**

- `hover`: show tooltip with explanation (from explanation system)
- `click`: expand detail panel below topology showing full technical details + diagnostic info

**Step 6: Verify topology renders**

Run server, check that SVG topology appears with nodes, connections, and animations.
Expected: 3-layer diagram with flowing connection lines

**Step 7: Commit**

```bash
git add cmd/server/static/index.html
git commit -m "feat: implement SVG layered topology with animations and interactions"
```

---

### Task 6: Frontend — Enhanced Diagnosis Display

**Files:**
- Modify: `cmd/server/static/index.html` (JS section)

**Step 1: Rewrite renderDiag with 3-part format**

Update diagnosis rendering to use: 现象 → 意味着什么 → 建议 format.

**Step 2: Add diagnosis panel to overview page**

Show diagnosis findings directly on the 网络全景 tab (below topology), not in a separate tab. Only show if there are findings.

**Step 3: Add first-time guidance hint**

Add a dismissable hint: "👆 点击拓扑图上的任意节点，了解它的作用"
Store dismissal state in `localStorage`.

**Step 4: Verify**

Expected: Diagnosis findings appear on overview page with rich explanations

**Step 5: Commit**

```bash
git add cmd/server/static/index.html
git commit -m "feat: enhance diagnosis display with 3-part format and topology integration"
```

---

### Task 7: Frontend — Traceroute Enhancement

**Files:**
- Modify: `cmd/server/static/index.html` (JS + HTML section)

**Step 1: Add preset target buttons**

Add 3 preset buttons above the input field:
- 测试百度 (baidu.com)
- 测试谷歌 (google.com)  
- 测试 DNS (8.8.8.8)

Each button calls `doTrace()` with the preset target.

**Step 2: Add hop zone classification**

Write `classifyHop(hop, index, totalHops)`:
- Zone 1 (local): private IP or first 2 hops → cyan
- Zone 2 (backbone): middle hops with public IP → orange
- Zone 3 (target): last 2 hops → purple

**Step 3: Add per-hop explanations**

- First hop: "📡 这是你的路由器"
- Timeout hops: "ℹ️ 这很正常！很多路由器拒绝回复探测包"
- Middle hops: show geo location if available
- RTT color coding: green <20ms, yellow 20-100ms, red >100ms

**Step 4: Add trace summary**

After trace completes, show summary line:
"你的数据到达 X 经过了 N 个节点，总延迟约 Xms（评价），途经 X 网络"

**Step 5: Verify traceroute**

Run server, test traceroute with preset buttons.
Expected: Preset buttons work, hops show zone colors and explanations

**Step 6: Commit**

```bash
git add cmd/server/static/index.html
git commit -m "feat: enhance traceroute with presets, zone classification, and explanations"
```

---

### Task 8: Frontend — Detail Info Tab (Interfaces + Routes + DNS)

**Files:**
- Modify: `cmd/server/static/index.html`

**Step 1: Consolidate interfaces + routes + DNS into 详细信息 tab**

Merge the old interfaces, routes, and DNS sections into a single tab with sub-sections. Each interface item uses the explanation system to show friendly labels.

**Step 2: Add expandable technical details**

Each interface shows friendly name + explanation by default. Click to expand shows MAC, MTU, IPv6, metric, etc.

**Step 3: Verify**

Expected: 详细信息 tab shows all technical info with friendly labels

**Step 4: Commit**

```bash
git add cmd/server/static/index.html
git commit -m "feat: consolidate detail info tab with expandable technical details"
```

---

### Task 9: Build & Smoke Test

**Files:** None (verification only)

**Step 1: Full build**

```bash
cd /home/hyh/Projects/routepeek
go build -o routepeek ./cmd/cli
go build -o routepeek-web ./cmd/server
```

Expected: Both binaries build without errors

**Step 2: CLI smoke test**

```bash
./routepeek scan
./routepeek diag
```

Expected: CLI output unchanged, new diagnostics may appear

**Step 3: Web UI smoke test**

```bash
./routepeek-web
```

Open http://localhost:8080 and verify:
- [ ] Status summary bar shows natural language description
- [ ] SVG topology renders with 3 layers
- [ ] Nodes are clickable with explanations
- [ ] Diagnostic issues marked on topology nodes
- [ ] 3-tab navigation works
- [ ] Traceroute presets work
- [ ] Detail info tab shows all technical info

**Step 4: Commit built binaries (optional)**

```bash
git add -A
git commit -m "feat: RoutePeek v2.0 - beginner-friendly UX enhancement"
```
