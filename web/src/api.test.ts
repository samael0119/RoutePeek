import { afterEach, describe, expect, it, vi } from 'vitest';
import { fetchConnectivity, fetchOverview, fetchTrace } from './api';

describe('api client', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('normalizes nullable snapshot collections from older servers', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            snapshot: {
              timestamp: '2026-04-30T00:00:00Z',
              interfaces: null,
              routes: null,
              dns: { servers: null, search: null, port: 0, is_secure: false },
              proxy: { http_proxy: '', https_proxy: '', no_proxy: '', has_proxy: false },
              vm_networks: null,
              default_gateway: '',
              public_ip: ''
            },
            diagnosis: { timestamp: '2026-04-30T00:00:00Z', findings: null, summary: '' },
            health: { status: 'ok', risk_level: 'low', label: '', summary: '' },
            actions: null,
            topology: { nodes: null, links: null, highlight_nodes: null }
          })
        )
      )
    );

    const overview = await fetchOverview('zh');

    expect(overview.snapshot.interfaces).toEqual([]);
    expect(overview.snapshot.routes).toEqual([]);
    expect(overview.snapshot.dns.servers).toEqual([]);
    expect(overview.snapshot.dns.search).toEqual([]);
    expect(overview.snapshot.vm_networks).toEqual([]);
    expect(overview.diagnosis.findings).toEqual([]);
    expect(overview.actions).toEqual([]);
    expect(overview.topology.nodes).toEqual([]);
    expect(overview.topology.links).toEqual([]);
    expect(overview.topology.highlight_nodes).toEqual([]);
  });

  it('normalizes nullable trace hops from restricted traceroute tools', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('null')));

    await expect(fetchTrace('example.com', 'zh')).resolves.toEqual([]);
  });

  it('normalizes nullable connectivity targets', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            timestamp: '2026-04-30T00:00:00Z',
            duration_ms: 0,
            summary: '',
            targets: null
          })
        )
      )
    );

    const report = await fetchConnectivity('zh');

    expect(report.targets).toEqual([]);
  });
});
