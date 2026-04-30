import type { APIErrorPayload, ConnectivityReport, OverviewResponse, TraceHop } from './types';
import { type Locale, tFor } from './i18n';

async function readJSON<T>(response: Response, locale: Locale = 'zh'): Promise<T> {
  const payload = (await response.json()) as T & APIErrorPayload;
  if (!response.ok) {
    throw new Error(payload.error?.message || tFor(locale, 'requestFailed', { status: response.status }));
  }
  return payload;
}

export async function fetchOverview(locale: Locale = 'zh'): Promise<OverviewResponse> {
  const response = await fetch(`/api/overview?lang=${encodeURIComponent(locale)}`);
  return normalizeOverview(await readJSON<OverviewResponse>(response, locale));
}

export async function fetchTrace(target: string, locale: Locale = 'zh'): Promise<TraceHop[]> {
  const response = await fetch(`/api/trace?target=${encodeURIComponent(target)}`);
  return (await readJSON<TraceHop[] | null>(response, locale)) ?? [];
}

export async function fetchConnectivity(locale: Locale = 'zh'): Promise<ConnectivityReport> {
  const response = await fetch(`/api/connectivity?lang=${encodeURIComponent(locale)}`);
  return normalizeConnectivity(await readJSON<ConnectivityReport>(response, locale));
}

export async function fetchReportMarkdown(locale: Locale = 'zh'): Promise<string> {
  const response = await fetch(`/api/report?format=markdown&lang=${encodeURIComponent(locale)}`);
  if (!response.ok) {
    const payload = (await response.json()) as APIErrorPayload;
    throw new Error(payload.error?.message || tFor(locale, 'requestFailed', { status: response.status }));
  }
  return response.text();
}

function normalizeOverview(overview: OverviewResponse): OverviewResponse {
  const diagnosis = overview.diagnosis ?? { timestamp: '', findings: [], summary: '' };
  const topology = overview.topology ?? { nodes: [], links: [], highlight_nodes: [] };
  const actions = (overview.actions ?? []).map((action) => ({
    ...action,
    steps: action.steps ?? []
  }));

  if (!overview.snapshot) {
    return {
      ...overview,
      diagnosis: {
        ...diagnosis,
        findings: diagnosis.findings ?? []
      },
      actions,
      topology: {
        ...topology,
        nodes: topology.nodes ?? [],
        links: topology.links ?? [],
        highlight_nodes: topology.highlight_nodes ?? []
      }
    };
  }
  const dns = overview.snapshot.dns ?? { servers: [], search: [], port: 0, is_secure: false };

  return {
    ...overview,
    diagnosis: {
      ...diagnosis,
      findings: diagnosis.findings ?? []
    },
    actions,
    topology: {
      ...topology,
      nodes: topology.nodes ?? [],
      links: topology.links ?? [],
      highlight_nodes: topology.highlight_nodes ?? []
    },
    snapshot: {
      ...overview.snapshot,
      interfaces: overview.snapshot.interfaces ?? [],
      routes: overview.snapshot.routes ?? [],
      dns: {
        ...dns,
        servers: dns.servers ?? [],
        search: dns.search ?? []
      },
      vm_networks: overview.snapshot.vm_networks ?? []
    }
  };
}

function normalizeConnectivity(report: ConnectivityReport): ConnectivityReport {
  return {
    ...report,
    targets: report.targets ?? []
  };
}
