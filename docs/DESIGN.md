# RoutePeek 项目设计文档

> 网络拓扑可视化与诊断工具 · Network Topology Visualization & Diagnostic Tool
> 版本: v1.0 | 更新: 2026-04-23

---

## 1. 项目概述

### 1.1 核心定位

RoutePeek 是一款面向**网络小白**的跨平台网络配置可视化工具，帮助用户理解多网卡、VPN、代理、虚拟机等复杂网络环境。

**目标用户**：
- 不熟悉网络的普通用户
- 需要同时管理有线/WiFi/VPN/虚拟机的开发者
- 需要排查「VPN 影响虚拟机连接」等混合网络问题的运维人员

**核心价值**：同时展示所有网络层（有线/WiFi/VPN/代理/虚拟机），并用通俗中文解释每项配置的含义。

### 1.2 核心功能

| 功能 | 说明 |
|------|------|
| 多网卡检测 | 同时展示有线、WiFi、VPN、虚拟机网络等所有网卡 |
| 路由路径可视化 | 纵向拓扑展示本机到公网的完整路径 |
| VPN/代理感知 | 检测 cnem/tun/tap/wg 等 VPN 接口；解析 NO_PROXY/HTTP_PROXY |
| DNS 诊断 | 识别非标准DNS，防范DNS劫持 |
| IP 归属地 | 公网IP、网关、DNS 服务器地理位置展示 |
| 自动诊断 | 一键检测常见网络问题并给出修复建议 |
| Traceroute | /api/trace?target= 追踪到任意目标的路径 |

---

## 2. 系统架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     用户交互层                                │
│     ┌─────────────────────┐    ┌─────────────────────┐     │
│     │   CLI (cmd/cli)     │    │  Web UI (cmd/server) │     │
│     │  scan / diag /      │    │  http://localhost   │     │
│     │  routes / interfaces│    │  :8080              │     │
│     │  (Cobra)            │    │  (vanilla HTML/JS)   │     │
│     └─────────┬───────────┘    └──────────┬──────────┘     │
│               │                           │                 │
│               └──────────┬─────────────────┘                 │
│                          ▼                                   │
│              ┌─────────────────────┐                        │
│              │  API 聚合层           │                        │
│              │  internal/discovery  │                        │
│              │  GetNetworkSnapshot()│                        │
│              │  RunDiagnostics()    │                        │
│              └──────────┬───────────┘                        │
│                          │                                   │
│          ┌───────────────┼───────────────┐                 │
│          ▼               ▼               ▼                 │
│  ┌────────────┐  ┌────────────┐  ┌────────────────┐        │
│  │ pkg/types  │  │ pkg/netinfo│  │ internal/      │        │
│  │            │  │            │  │ diagnostic     │        │
│  │ 数据结构    │  │ 网络信息采集 │  │ 诊断引擎       │        │
│  └────────────┘  └────────────┘  └────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 模块职责

| 模块 | 路径 | 职责 |
|------|------|------|
| **cmd/cli** | `cmd/cli/main.go` | Cobra CLI 入口，提供 scan/diag/routes/interfaces 命令 |
| **cmd/server** | `cmd/server/main.go` | Gorilla/Mux HTTP 服务器，静态资源嵌入 Go 二进制 |
| **internal/discovery** | `internal/discovery/snapshot.go` | 聚合所有网络信息为 NetworkSnapshot |
| **internal/diagnostic** | `internal/diagnostic/diagnose.go` | 诊断引擎，检查 DNS/路由/VPN/代理/连通性问题 |
| **pkg/netinfo** | `pkg/netinfo/*.go` | 跨平台网络信息采集（纯 Go，无 cgo） |
| **pkg/types** | `pkg/types/types.go` | 所有数据结构定义 |

---

## 3. 详细设计

### 3.1 pkg/netinfo — 跨平台网络信息采集

纯 Go 实现，无 cgo 依赖，便于交叉编译。通过执行平台特定 Shell 命令解析结果。

#### 3.1.1 interface.go — 网卡枚举

```
GetNetworkInterfaces() []NetworkInterface
```

- Linux: `ip -br addr` + `ip link`
- macOS: `ifconfig`
- Windows: `getmac /v` + `ipconfig`

**NetworkInterface 字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | 接口名（如 ens38, docker0, cnem_vnic） |
| Index | int | 内核接口索引 |
| MTU | int | 最大传输单元 |
| MAC | string | MAC 地址 |
| IP4 / IP6 | string | IPv4 / IPv6 地址 |
| IsUp | bool | 是否启用 |
| IsLoopback | bool | 是否 loopback |
| Type | string | 类型：ethernet/wifi/vpn/vm/loopback/unknown |

**Type 判断逻辑**（interface.go:82）：
- `name == "lo"` → loopback
- `name` 含 docker / bridge → vm
- `name` 含 tun/tap/wg/utun/ipsec/cnem/cnem_vnic → vpn
- 默认 → ethernet

#### 3.1.2 route.go — 路由表

```
GetRoutes() []RouteEntry
GetDefaultGateway() (string, error)
IsVPNGlobal(vpnInterface string, routes []RouteEntry) bool
```

- Linux: `ip route show`
- macOS: `netstat -rn`
- Windows: `route print`

**RouteEntry 字段**：`Destination`, `Gateway`, `Interface`, `Metric`, `Table`

#### 3.1.3 dns.go — DNS 配置

```
GetDNSConfig() (*DNSConfig, error)
```

- Linux: 读取 `/etc/resolv.conf`
- macOS: `scutil --dns`
- Windows: `ipconfig /all`

**已知公共 DNS 白名单**：
- Google: 8.8.8.8, 8.8.4.4, 2001:4860:4860::8888
- Cloudflare: 1.1.1.1, 1.0.0.1, 2606:4700:4700::1111
- Quad9: 9.9.9.9, 149.112.112.112
- OpenDNS: 208.67.222.222, 208.67.220.220
- Ali DNS: 223.5.5.5, 223.6.6.6
- 114 DNS: 114.114.114.114, 114.114.115.115

#### 3.1.4 proxy.go — 代理检测

```
DetectProxy() *ProxyConfig
```

读取环境变量：`HTTP_PROXY`, `http_proxy`, `HTTPS_PROXY`, `https_proxy`, `ALL_PROXY`, `all_proxy`, `NO_PROXY`, `no_proxy`

Windows 额外检查：`netsh winhttp show proxy`

**注意**：Clash 等代理工具通常只修改 `NO_PROXY` 而不设置 `HTTP_PROXY`，因此 `has_proxy` 为 true 但 `http_proxy` 可能为空字符串。

#### 3.1.5 vpn.go — VPN 检测

```
DetectVPN() *VPNInfo
DetectVMNetworks() []VMNetwork
```

**VPN 检测**：扫描活跃网卡名是否匹配前缀：
```
tun, tap, ppp, vpnc, vpn, wg, utun, ipsec, cnem, cnem_vnic
```

**VPNInfo 字段**：`Status`(connected/disconnected/unknown), `Interface`, `ClientIP`, `Protocol`, `ServerIP`, `IsGlobal`, `Name`

**VPN 协议推断**（按 IP 段）：
| IP 段 | 协议 |
|-------|------|
| 10.0.0.0/8 | openvpn/wireguard |
| 172.16.0.0/12 | ipsec/vpn |
| 192.168.0.0/16 | various |

#### 3.1.6 consts.go — 常量

```go
VPNInterfacePrefixes = []string{"tun","tap","ppp","vpnc","vpn","wg","utun","ipsec","cnem","cnem_vnic"}

CommonDockerSubnets = []string{"172.17.0.0/16","172.18.0.0/16",...}
CommonVMwareSubnets = []string{"192.168.56.0/24","192.168.127.0/24",...}
KnownVMSubnets = []struct{Name, Type, Subnet}{...}
```

### 3.2 pkg/types — 数据结构

```go
type NetworkInterface struct {
    Name, MAC, IP4, IP6, Type string
    Index, MTU                int
    IsUp, IsLoopback          bool
}

type RouteEntry struct {
    Destination, Gateway, Interface, Table string
    Metric                                 int
}

type DNSConfig struct {
    Servers []string
    Search []string
    Port   int
    IsSecure bool  // 是否使用 DNS over HTTPS/TLS
}

type ProxyConfig struct {
    HTTPProxy, HTTPSProxy, NoProxy string
    HasProxy bool
}

type VPNInfo struct {
    Name, Status, ServerIP, ClientIP, Protocol, Interface string
    IsGlobal bool
}

type VMNetwork struct {
    Name, Type, Subnet, Gateway, HostIP string
    DHCP bool
}

type NetworkSnapshot struct {
    Timestamp      time.Time
    Interfaces     []NetworkInterface
    Routes         []RouteEntry
    DNS            DNSConfig
    Proxy          ProxyConfig
    VPN            *VPNInfo
    VMNetworks     []VMNetwork
    DefaultGateway string
    PublicIP       string
}

type TraceHop struct {
    Hop, Address, RTT1, RTT2, RTT3 string
    Hostname string
}

type DiagnosisReport struct {
    Timestamp time.Time
    Findings  []DiagnosisResult
    Summary   string
}

type DiagnosisResult struct {
    Severity, Code, Title, Message, Suggestion string
}
```

### 3.3 internal/discovery — 信息聚合

```
GetNetworkSnapshot() (*NetworkSnapshot, error)
```

按顺序调用 netinfo 各模块，组装 NetworkSnapshot。

另外通过 HTTP 请求外部服务获取公网 IP：
- https://api.ipify.org?format=text
- https://ifconfig.me/ip
- https://icanhazip.com

（按顺序尝试，任一成功即返回）

### 3.4 internal/diagnostic — 诊断引擎

```
RunDiagnostics(snapshot *types.NetworkSnapshot) *types.DiagnosisReport
```

#### 诊断项目

| 诊断项 | Code | Severity | 说明 |
|--------|------|----------|------|
| 默认网关不可达 | DEFAULT_GW_UNREACHABLE | warning | 默认网关 IP 不通 |
| DNS 被篡改 | SUSPICIOUS_DNS | warning | DNS 不在白名单中 |
| VPN 影响 LAN | VPN_LAN_IMPACT | warning | 0.0.0.0/0 路由覆盖虚拟机网段 |
| 虚拟机网段不可达 | VM_UNREACHABLE | info | Docker/VMware 网段连通性失败 |
| 代理直连 | PROXY_DIRECT | info | 代理已启用但直连目标 |
| 路由 metric 冲突 | METRIC_CONFLICT | warning | 同一目标多条路由 metric 不同 |

### 3.5 cmd/server — Web 服务器

#### 3.5.1 架构

- **HTTP 路由**：Gorilla/Mux
- **静态资源**：`//go:embed static/*` 嵌入 Go 二进制，无需独立文件部署
- **API**：RESTful JSON，无鉴权（本地工具）

#### 3.5.2 API 端点

| 端点 | 方法 | 说明 | 响应 |
|------|------|------|------|
| `/api/snapshot` | GET | 完整网络快照 | NetworkSnapshot JSON |
| `/api/diagnosis` | GET | 网络诊断报告 | DiagnosisReport JSON |
| `/api/interfaces` | GET | 仅网卡列表 | []NetworkInterface JSON |
| `/api/routes` | GET | 仅路由表 | []RouteEntry JSON |
| `/api/trace?target=` | GET | Traceroute 到目标 | []TraceHop JSON |
| `/` | GET | Web UI | index.html |

#### 3.5.3 Web UI（index.html）

- **框架**：原生 HTML + CSS + JavaScript，无框架依赖
- **字体**：JetBrains Mono（等宽）/ Inter（UI）
- **样式**：深色赛博朋克主题，CSS 变量驱动配色
- **数据获取**：原生 `fetch()` 调用 API，页面内无构建工具

**主要页面区块**：

```
┌──────────────────────────────────────┐
│  ROUTEPEEK (header)                   │
├──────────────────────────────────────┤
│  [概览] [诊断] [路由] [接口] [追踪]     │  ← Nav tabs
├──────────────────────────────────────┤
│  最后更新: 2026-04-23 14:38:45        │
│  [扫描刷新 button]                    │  ← action-bar
├──────────────────────────────────────┤
│  ┌─ 网络路径拓扑 ──────────────────┐  │
│  │ 🖥️ 本机 localhost               │  │
│  │ › ⚡ 网关 192.168.1.1 [归属地]   │  │
│  │ › 🔒 VPN 192.168.100.23 [cnem]  │  │
│  │ › 🌐 代理 127.0.0.1:7897        │  │
│  │ › 🔤 DNS 10.10.3.197 [归属地]   │  │
│  │ › 🌐 公网出口 1.198.18.94 [归属]│  │
│  └──────────────────────────────────┘  │
│  ┌─ 网络统计 ─────────────────────┐   │
│  │ 📡 活跃接口   📋 路由条目        │   │
│  │ 🔤 DNS服务器  🔒 VPN状态         │   │
│  └──────────────────────────────────┘   │
└──────────────────────────────────────┘
```

**CSS 变量（主题色）**：

```css
--bg-void:      #020408   /* 最深背景 */
--bg-deep:      #040a14
--cyan:         #00d4ff   /* 主色调 */
--green:        #00ff94   /* 成功/DNS */
--red:          #ff3a5c   /* VPN/错误 */
--orange:       #ff9f1a   /* 网关 */
--purple:       #b44fff   /* 公网 */
```

**拓扑节点类型与配色**：

| 节点类型 | class | 主色 |
|----------|-------|------|
| 本机 | 默认 | cyan |
| 网关 | `.gateway` | orange |
| VPN | `.vpn` | red |
| DNS | `.dns-node` | green |
| 公网 | `.public` | purple |
| 代理 | `.proxy` | cyan |

**IP 归属地**：浏览器端调用 `ip-api.com`（免费 API，45 req/min），结果缓存于 `geoCache` 对象。私网 IP（10.x、172.16-31.x、192.168.x）跳过查询。

### 3.6 cmd/cli — 命令行工具

基于 Cobra 框架，提供 4 个子命令：

| 命令 | 说明 |
|------|------|
| `routepeek scan` | 完整网络配置概览（默认命令） |
| `routepeek diag` | 运行诊断并输出报告 |
| `routepeek routes` | 仅显示路由表 |
| `routepeek interfaces` | 仅显示网卡列表 |

---

## 4. 关键设计决策

### 4.1 纯 Go，无 cgo

所有网络信息通过执行 Shell 命令（`ip`、`netstat`、`ifconfig` 等）并解析输出来获取，无需调用系统 API（避免 cgo 依赖，便于 `GOOS=linux GOARCH=arm64 go build` 交叉编译）。

### 4.2 诊断而非修复

Phase 1 只做检测和报告，修复功能（Phase 2）待定。诊断结果附带 `Suggestion` 字段，给出具体修复建议而非纯警告。

### 4.3 静态嵌入二进制

`//go:embed static/*` 将 Web UI 嵌入 Go 二进制，部署时只需一个可执行文件，无需解压 HTML/JS/CSS 资源目录。

### 4.4 前端无框架

Web UI 使用原生 HTML/JS/CSS，不引入 React/Vue 等框架。这样：
- 无需 Node.js 构建工具链
- `go build` 即可完成全部构建
- 二进制体积可控（当前 ~10MB）

代价是大型前端项目难以维护，可在后续迁移 Vue/React（参见 TODO）。

### 4.5 归属地查询在浏览器端

GeoIP 查询由浏览器 JS 调用 `ip-api.com`，而非 Go 服务端。这样：
- 绕过代理限制（Go 服务端可能走代理，无法直接查询公网 IP）
- 减少服务端依赖
- 用户浏览器有直接网络访问时查询成功率高

---

## 5. 项目文件结构

```
RoutePeek/
├── cmd/
│   ├── cli/
│   │   └── main.go          # Cobra CLI 入口 (114 行)
│   └── server/
│       ├── main.go          # HTTP 服务器入口 (161 行)
│       └── static/
│           └── index.html   # Web UI (1385 行，嵌入二进制)
├── internal/
│   ├── diagnostic/
│   │   ├── diagnose.go      # 诊断引擎核心 (272 行)
│   │   └── print.go         # 诊断报告格式化 (68 行)
│   └── discovery/
│       ├── snapshot.go      # 信息聚合 (110 行)
│       └── print.go         # CLI 格式化输出 (204 行)
├── pkg/
│   ├── netinfo/             # 跨平台网络信息采集 (974 行)
│   │   ├── consts.go        # 常量定义 (54 行)
│   │   ├── dns.go           # DNS 配置 (147 行)
│   │   ├── interface.go     # 网卡枚举 (142 行)
│   │   ├── proxy.go         # 代理检测 (46 行)
│   │   ├── route.go         # 路由表 (474 行)
│   │   └── vpn.go           # VPN 检测 (111 行)
│   └── types/
│       └── types.go         # 数据结构 (103 行)
├── docs/
│   └── DESIGN.md            # 本文档
├── go.mod
├── go.sum
└── README.md
```

**代码量统计**：~1845 行 Go + 1385 行 HTML/JS/CSS

---

## 6. 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | ≥1.18 |
| HTTP 路由 | gorilla/mux | v1.8.1 |
| CLI 框架 | spf13/cobra | v1.8.0 |
| 静态嵌入 | go:embed | (标准库) |
| 前端 | Vanilla HTML/CSS/JS | — |
| 字体 | Google Fonts (JetBrains Mono, Inter) | — |
| 归属地 API | ip-api.com | 免费层 |

---

## 7. 构建与运行

```bash
# 编译 CLI
go build -o routepeek ./cmd/cli

# 编译 Web Server
go build -o routepeek-web ./cmd/server

# 运行 CLI
./routepeek scan
./routepeek diag
./routepeek routes
./routepeek interfaces

# 运行 Web
./routepeek-web
# → http://localhost:8080
```

---

## 8. 待完成（TODO）

- [ ] Phase 2: 一键修复（备份 + 确认 + 回滚）
- [ ] Web UI 迁移 Vue/React（长期维护）
- [ ] Web UI 改用 SVG 可视化拓扑图
- [ ] 创建 Makefile 支持跨平台编译
- [ ] 支持 DNS over HTTPS/TLS 检测（IsSecure 字段）
- [ ] Traceroute 跳跃点地理定位
