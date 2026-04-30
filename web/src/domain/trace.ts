import type { TraceHop } from '../types';
import { type Locale, tFor } from '../i18n';

export interface TraceExplanation {
  state: 'empty' | 'all_timeout' | 'reachable_with_middle_timeout' | 'reachable' | 'partial' | 'route_probe' | 'proxy_fake_ip';
  mode: 'traceroute' | 'route_probe' | 'proxy_fake_ip';
  panelTitle: string;
  title: string;
  message: string;
  hopCount: number;
}

export function explainTrace(hops: TraceHop[], locale: Locale = 'zh'): TraceExplanation {
  if (hops.length === 0) {
    return {
      state: 'empty',
      mode: 'traceroute',
      panelTitle: tFor(locale, 'traceTitle'),
      title: tFor(locale, 'traceEmptyTitle'),
      message: tFor(locale, 'traceEmptyMessage'),
      hopCount: 0
    };
  }

  const mode = traceMode(hops);
  if (mode === 'route_probe') {
    return {
      state: 'route_probe',
      mode,
      panelTitle: tFor(locale, 'traceRouteProbePanelTitle'),
      title: tFor(locale, 'traceRouteProbeTitle'),
      message: tFor(locale, 'traceRouteProbeMessage'),
      hopCount: hops.length
    };
  }
  if (mode === 'proxy_fake_ip') {
    return {
      state: 'proxy_fake_ip',
      mode,
      panelTitle: tFor(locale, 'traceProxyFakeIpPanelTitle'),
      title: tFor(locale, 'traceProxyFakeIpTitle'),
      message: tFor(locale, 'traceProxyFakeIpMessage'),
      hopCount: hops.length
    };
  }

  const timeoutFlags = hops.map(isTimeoutHop);
  const allTimeout = timeoutFlags.every(Boolean);
  const targetReachable = !timeoutFlags[timeoutFlags.length - 1];
  const hasMiddleTimeout = timeoutFlags.slice(0, -1).some(Boolean);

  if (allTimeout) {
    return {
      state: 'all_timeout',
      mode,
      panelTitle: tFor(locale, 'traceTitle'),
      title: tFor(locale, 'traceAllTimeoutTitle'),
      message: tFor(locale, 'traceAllTimeoutMessage'),
      hopCount: hops.length
    };
  }

  if (targetReachable && hasMiddleTimeout) {
    return {
      state: 'reachable_with_middle_timeout',
      mode,
      panelTitle: tFor(locale, 'traceTitle'),
      title: tFor(locale, 'traceMiddleTimeoutTitle'),
      message: tFor(locale, 'traceMiddleTimeoutMessage'),
      hopCount: hops.length
    };
  }

  if (targetReachable) {
    return {
      state: 'reachable',
      mode,
      panelTitle: tFor(locale, 'traceTitle'),
      title: tFor(locale, 'traceReachableTitle', { count: hops.length }),
      message: tFor(locale, 'traceReachableMessage'),
      hopCount: hops.length
    };
  }

  return {
    state: 'partial',
    mode,
    panelTitle: tFor(locale, 'traceTitle'),
    title: tFor(locale, 'tracePartialTitle'),
    message: tFor(locale, 'tracePartialMessage'),
    hopCount: hops.length
  };
}

export function traceMode(hops: TraceHop[]): 'traceroute' | 'route_probe' | 'proxy_fake_ip' {
  if (hops.some((hop) => hop.mode === 'route_probe')) {
    return 'route_probe';
  }
  if (hops.some((hop) => hop.mode === 'proxy_fake_ip')) {
    return 'proxy_fake_ip';
  }
  return 'traceroute';
}

export function isTimeoutHop(hop: TraceHop): boolean {
  const values = [hop.address, hop.rtt1, hop.rtt2, hop.rtt3].map((value) => value.trim());
  return values.every((value) => value === '' || value === '*');
}
