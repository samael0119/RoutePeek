export interface NetworkInterface {
  name: string;
  index: number;
  mtu: number;
  mac: string;
  ip4: string;
  ip6: string;
  is_up: boolean;
  is_loopback: boolean;
  type: string;
}

export interface RouteEntry {
  destination: string;
  gateway: string;
  interface: string;
  metric: number;
  table: string;
}

export interface DNSConfig {
  servers: string[];
  search: string[];
  port: number;
  is_secure: boolean;
}

export interface ProxyConfig {
  http_proxy: string;
  https_proxy: string;
  no_proxy: string;
  has_proxy: boolean;
}

export interface VPNInfo {
  name: string;
  status: string;
  server_ip: string;
  client_ip: string;
  protocol: string;
  interface: string;
  is_global: boolean;
}

export interface VMNetwork {
  name: string;
  type: string;
  subnet: string;
  gateway: string;
  dhcp: boolean;
  host_ip: string;
}

export interface NetworkSnapshot {
  timestamp: string;
  interfaces: NetworkInterface[];
  routes: RouteEntry[];
  dns: DNSConfig;
  proxy: ProxyConfig;
  vpn?: VPNInfo;
  vm_networks: VMNetwork[];
  default_gateway: string;
  public_ip: string;
}

export interface DiagnosisResult {
  severity: 'critical' | 'warning' | 'info' | string;
  code: string;
  title: string;
  message: string;
  suggestion: string;
}

export interface DiagnosisReport {
  timestamp: string;
  findings: DiagnosisResult[];
  summary: string;
}

export interface HealthOverview {
  status: 'ok' | 'attention' | 'offline' | string;
  risk_level: 'low' | 'medium' | 'high' | string;
  label: string;
  primary_issue?: string;
  summary: string;
}

export interface ActionItem {
  code: string;
  severity: 'critical' | 'warning' | 'info' | string;
  title: string;
  category: string;
  impact: string;
  steps: string[];
  verify: string;
  target_node: string;
  confidence: 'high' | 'medium' | 'low' | string;
}

export interface TopologyNode {
  id: string;
  label: string;
  kind: string;
  status: string;
  description?: string;
}

export interface TopologyLink {
  from: string;
  to: string;
  status: string;
  label?: string;
}

export interface NetworkTopology {
  nodes: TopologyNode[];
  links: TopologyLink[];
  highlight_nodes: string[];
}

export interface OverviewResponse {
  snapshot: NetworkSnapshot;
  diagnosis: DiagnosisReport;
  health: HealthOverview;
  actions: ActionItem[];
  topology: NetworkTopology;
}

export interface ConnectivityTarget {
  name: string;
  address: string;
  kind: string;
}

export interface ConnectivityProbeResult {
  status: 'ok' | 'warning' | 'danger' | 'info' | string;
  code?: string;
  detail: string;
  endpoint?: string;
  latency_ms?: number;
}

export interface ConnectivityTargetResult {
  target: ConnectivityTarget;
  dns: ConnectivityProbeResult;
  tcp: ConnectivityProbeResult;
  http: ConnectivityProbeResult;
  path: ConnectivityProbeResult;
  severity: 'ok' | 'warning' | 'danger' | 'info' | string;
  conclusion: string;
}

export interface ConnectivityReport {
  timestamp: string;
  duration_ms: number;
  summary: string;
  targets: ConnectivityTargetResult[];
}

export interface TraceHop {
  hop: number;
  address: string;
  rtt1: string;
  rtt2: string;
  rtt3: string;
  hostname?: string;
  mode?: 'traceroute' | 'route_probe' | 'proxy_fake_ip' | string;
}

export interface APIErrorPayload {
  error?: {
    code: string;
    message: string;
  };
}
