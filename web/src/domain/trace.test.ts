import { describe, expect, it } from 'vitest';
import { explainTrace } from './trace';
import type { TraceHop } from '../types';

describe('trace explanations', () => {
  it('explains all timeout results without claiming the site is down', () => {
    const hops = [
      { hop: 1, address: '*', rtt1: '*', rtt2: '*', rtt3: '*' },
      { hop: 2, address: '*', rtt1: '*', rtt2: '*', rtt3: '*' }
    ] as TraceHop[];

    const explanation = explainTrace(hops, 'en');

    expect(explanation.state).toBe('all_timeout');
    expect(explanation.title).toBe('All hops timed out');
  });

  it('explains middle timeout as usually non-blocking when target is reachable', () => {
    const hops = [
      { hop: 1, address: '192.168.1.1', rtt1: '1ms', rtt2: '1ms', rtt3: '2ms' },
      { hop: 2, address: '*', rtt1: '*', rtt2: '*', rtt3: '*' },
      { hop: 3, address: '8.8.8.8', rtt1: '20ms', rtt2: '21ms', rtt3: '20ms' }
    ] as TraceHop[];

    const explanation = explainTrace(hops, 'en');

    expect(explanation.state).toBe('reachable_with_middle_timeout');
    expect(explanation.message).toContain('usually');
  });

  it('reports reachable target and hop count', () => {
    const hops = [
      { hop: 1, address: '192.168.1.1', rtt1: '1ms', rtt2: '1ms', rtt3: '2ms' },
      { hop: 2, address: '8.8.8.8', rtt1: '20ms', rtt2: '21ms', rtt3: '20ms' }
    ] as TraceHop[];

    const explanation = explainTrace(hops, 'en');

    expect(explanation.state).toBe('reachable');
    expect(explanation.title).toBe('Target reachable in 2 hops');
    expect(explanation.hopCount).toBe(2);
  });

  it('explains route probe results as limited connectivity probing', () => {
    const hops = [
      { hop: 1, address: '198.18.0.2', rtt1: '', rtt2: '', rtt3: '', hostname: '198.18.0.2 via Mihomo', mode: 'route_probe' },
      { hop: 2, address: '39.156.66.10', rtt1: '35.2 ms', rtt2: '34.8 ms', rtt3: '36.1 ms', hostname: '39.156.66.10 tcp/443 reachable', mode: 'route_probe' }
    ] as TraceHop[];

    const explanation = explainTrace(hops, 'en');

    expect(explanation.mode).toBe('route_probe');
    expect(explanation.panelTitle).toBe('Exit and Target Probe');
    expect(explanation.title).toBe('Limited trace: showing exit and target reachability');
    expect(explanation.message).toContain('intermediate hops');
    expect(explanation.message).toContain('administrator');
  });

  it('explains proxy fake-ip results without claiming a public multi-hop path', () => {
    const hops = [
      { hop: 1, address: '198.18.0.29', rtt1: '0.765 ms', rtt2: '0.812 ms', rtt3: '0.779 ms', mode: 'proxy_fake_ip' }
    ] as TraceHop[];

    const explanation = explainTrace(hops, 'en');

    expect(explanation.mode).toBe('proxy_fake_ip');
    expect(explanation.panelTitle).toBe('Proxy Virtual Exit');
    expect(explanation.title).toBe('Proxy fake-IP route detected');
    expect(explanation.message).toContain('198.18.0.0/15');
  });
});
