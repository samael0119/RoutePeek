import type { ActionItem, HealthOverview } from '../types';
import { type Locale, tFor } from '../i18n';

export type HealthTone = 'ok' | 'warning' | 'danger';

const severityWeight: Record<string, number> = {
  critical: 0,
  warning: 1,
  info: 2
};

export function healthTone(health: HealthOverview): HealthTone {
  if (health.status === 'offline') {
    return 'danger';
  }
  if (health.status === 'attention') {
    return 'warning';
  }
  return 'ok';
}

export function primaryAction(actions: ActionItem[]): ActionItem | undefined {
  return [...actions].sort((a, b) => severityRank(a.severity) - severityRank(b.severity))[0];
}

export function severityRank(severity: string): number {
  return severityWeight[severity] ?? 3;
}

export function severityLabel(severity: string, locale: Locale = 'zh'): string {
  if (severity === 'critical') {
    return tFor(locale, 'severityCritical');
  }
  if (severity === 'warning') {
    return tFor(locale, 'severityWarning');
  }
  return tFor(locale, 'severityInfo');
}

export function categoryLabel(category: string, locale: Locale = 'zh'): string {
  const labels: Record<string, Parameters<typeof tFor>[1]> = {
    connectivity: 'categoryConnectivity',
    dns: 'categoryDns',
    vpn: 'categoryVpn',
    proxy: 'categoryProxy',
    route: 'categoryRoute',
    vm: 'categoryVm',
    public_ip: 'categoryPublicIp'
  };
  return tFor(locale, labels[category] ?? 'categoryNetwork');
}
