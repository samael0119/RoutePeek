# RoutePeek - 网络拓扑可视化与诊断工具

一款面向网络小白的跨平台网络配置可视化工具，帮助用户理解多网卡、VPN、代理、虚拟机等复杂网络环境。

## 核心特性

| 功能 | 说明 |
|------|------|
| 多网卡检测 | 同时展示有线、WiFi、VPN、虚拟机网络等所有网卡 |
| 路由路径可视化 | 展示本机到目标地址的完整路由路径 |
| VPN/代理感知 | 检测全局VPN对虚拟机网段的影响 |
| DNS 诊断 | 识别非标准DNS，防范DNS劫持 |
| 自动诊断 | 一键检测常见网络问题并给出修复建议 |

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

## 输出示例

```
╔══════════════════════════════════════════════════════════════╗
║                     RoutePeek - 网络配置总览                  ║
╚══════════════════════════════════════════════════════════════╝

┌─────────────────────────────────────────────────────────────────────┐
│ 📡 网络接口                                                             │
└─────────────────────────────────────────────────────────────────────┘
   🌐 ens33 ✅
      MAC:  00:0c:29:c0:77:49
      MTU:  1500
   🔌 ens38 ✅
      IP:   192.168.1.63
      MAC:  00:0c:29:c0:77:5d
      MTU:  1500
   🌐 docker0 ✅
      IP:   172.17.0.1
      MAC:  2e:ac:5d:f8:e4:1d
      MTU:  1500
```

```
╔══════════════════════════════════════════════════════════════╗
║                       网络诊断报告                              ║
╚══════════════════════════════════════════════════════════════╝

📊 诊断结果: 发现 1 个警告，建议检查。

🟡 [SUSPICIOUS_DNS] 检测到非标准 DNS 服务器
──────────────────────────────────────────────────────────────────────
   DNS 服务器 10.10.3.197 不是已知公共 DNS，可能存在 DNS 劫持风险。

   💡 建议: 如果这不是你配置的 DNS，建议更改为公共 DNS：8.8.8.8 (Google) 或 1.1.1.1 (Cloudflare)。
```

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
- [ ] Phase 3: Web UI 改用 SVG 可视化拓扑图
- [ ] 添加 GeoIP 支持（traceroute 跳跃点地理定位）
- [ ] 创建 Makefile 支持跨平台编译

## 设计背景

MTR/WinMTR 等工具面向高级用户，纯 CLI 输出对网络小白不友好。本工具的核心价值在于**同时展示所有网络层**（有线/WiFi/VPN/代理/虚拟机），并用通俗中文解释每项配置的含义，特别适合排查「VPN 影响虚拟机连接」等常见问题。
