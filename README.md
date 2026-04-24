# RoutePeek - 网络拓扑可视化与诊断工具

一款面向网络小白的跨平台网络配置可视化工具，帮助用户理解多网卡、VPN、代理、虚拟机等复杂网络环境。

## 核心特性

| 功能 | 说明 |
|------|------|
| 多网卡检测 | 同时展示有线、WiFi、VPN、虚拟机网络等所有网卡 |
| SVG 分层拓扑 | 图形化展示数据从本机到公网的完整路径（Web 端） |
| 通俗解释系统 | 提供网络术语的「大白话」解释，小白也能看懂 |
| 智能诊断联动 | 诊断结果直接在拓扑图上以脉冲高亮显示，直观定位问题 |
| VPN/代理感知 | 检测全局VPN、代理配置及其对局域网的影响 |
| Traceroute | `/api/trace?target=` 场景化追踪到任意目标的路径 |

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI / Web UI                        │
├─────────────────────────────────────────────────────────────┤
│  cmd/cli/main.go          │  cmd/server/main.go            │
│  - Cobra 命令行            │  - HTTP API + 静态页面          │
│  - scan/diag/routes        │  - Gorilla/Mux 路由            │
└─────────────┬───────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────┐
│                   internal/discovery                        │
│  - GetNetworkSnapshot()  聚合所有网络信息                    │
│  - RunDiagnostics()      执行诊断检查                        │
│  - PrintNetworkOverview() CLI 格式化输出                     │
└─────────────┬───────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────┐
│                     pkg/netinfo                              │
│  跨平台网络信息采集（纯 Go，无 cgo）                          │
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

## 技术方案

### 跨平台实现
- **Linux/macOS/Windows** 统一接口，底层通过 Shell 命令实现
- 避免 cgo 依赖，便于交叉编译
- 网络接口枚举：`net.Interfaces()` + 平台特定命令
- 路由表读取：`ip route` / `netstat -rn` / `route print`
- DNS 配置读取：`/etc/resolv.conf` / `scutil --dns` / `ipconfig /all`
- VPN 检测：扫描 `tun`/`tap`/`wg`/`utun` 前缀网卡 + 配置目录扫描

### 诊断引擎
- DNS 异常检测：比对已知公共 DNS 列表（Google/Cloudflare/Quad9/OpenDNS/阿里/114）
- 路由冲突检测：同一目标网段存在多条路由
- VPN LAN 影响检测：0.0.0.0/0 覆盖虚拟机网段（192.168.56.0/24、172.17.0.0/16 等）
- 虚拟机网段连通性检测：探测 Docker/VMware/VirtualBox/Hyper-V 网段

### 安全机制
- 一键修复（Phase 2）：备份 → 二次确认 → 自动回滚
- 诊断结果附带具体修复建议，非纯警告

## 快速开始

```bash
# 编译
go build -o RoutePeek ./cmd/cli
go build -o RoutePeek-web ./cmd/server

# 命令行使用
./RoutePeek scan          # 查看完整网络配置
./RoutePeek diag          # 运行网络诊断
./RoutePeek routes        # 查看路由表
./RoutePeek interfaces    # 查看网卡详情

# Web 界面
./RoutePeek-web           # 启动 Web 服务器（默认 8080 端口）
```

## 界面概览

### 1. 网络全景（主页）
- **状态摘要**：自然语言一句话总结当前网络状态（例：「✅ 你正通过有线网卡连接路由器上网」）
- **SVG 分层拓扑图**：动态绘制 LAN 层、网关层、外网层，节点间用流动动画连接
- **诊断面板**：如果存在网络异常（如 DNS 篡改、多网卡冲突），直接在下方展示原因和修复建议

### 2. 详细信息
供进阶用户查看完整的技术参数：
- 各网卡详细配置（MAC、MTU、状态）
- 系统路由表（目标网络、网关、Metric）
- DNS 服务器及代理设置

### 3. 路径追踪
- 提供预设场景（测试百度、测试谷歌、测试 DNS）
- 追踪结果分区域着色（本地网络、运营商骨干、目标服务器），带有每跳的通俗解释

## 项目结构

```
RoutePeek/
├── cmd/
│   ├── cli/main.go          # CLI 入口（Cobra）
│   └── server/main.go       # Web 服务器入口
├── internal/
│   ├── diagnostic/          # 诊断引擎
│   │   ├── diagnose.go      # RunDiagnostics() 核心逻辑
│   │   └── print.go         # 诊断报告格式化输出
│   └── discovery/           # 网络信息聚合
│       ├── snapshot.go      # GetNetworkSnapshot()
│       └── print.go         # CLI 格式化输出
├── pkg/
│   ├── netinfo/             # 跨平台网络信息采集
│   │   ├── interface.go     # 网卡枚举
│   │   ├── route.go         # 路由表
│   │   ├── dns.go           # DNS 配置
│   │   ├── proxy.go         # 代理检测
│   │   ├── vpn.go           # VPN/虚拟机网络
│   │   └── consts.go        # 常量定义
│   └── types/               # 数据结构
│       └── types.go
├── static/
│   └── index.html           # Web UI（可选）
├── go.mod
└── README.md
```

## 依赖

```
github.com/gorilla/mux v1.8.1    # HTTP 路由
github.com/spf13/cobra v1.8.0     # CLI 框架
```

## 后续计划

- [ ] Phase 2: 一键修复（备份 + 确认 + 回滚）
- [ ] Web UI 改用 SVG 可视化拓扑图
- [ ] Web UI 迁移 Vue/React（长期可维护性）
- [ ] 创建 Makefile 支持跨平台编译
- [ ] 支持 DNS over HTTPS/TLS 检测（IsSecure 字段）
- [ ] Traceroute 跳跃点地理定位

## 设计背景

MTR/WinMTR 等工具面向高级用户，纯 CLI 输出对网络小白不友好。本工具的核心价值在于**同时展示所有网络层**（有线/WiFi/VPN/代理/虚拟机），并用通俗中文解释每项配置的含义，特别适合排查「VPN 影响虚拟机连接」等常见问题。
