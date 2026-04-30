import { describe, expect, it } from 'vitest';
import { healthTone, primaryAction } from './overview';
import type { ActionItem, HealthOverview } from '../types';

describe('overview domain helpers', () => {
  it('maps health states to stable UI tones', () => {
    expect(healthTone({ status: 'ok' } as HealthOverview)).toBe('ok');
    expect(healthTone({ status: 'attention' } as HealthOverview)).toBe('warning');
    expect(healthTone({ status: 'offline' } as HealthOverview)).toBe('danger');
  });

  it('selects the highest severity action first', () => {
    const actions = [
      { code: 'MULTI_NIC_ACTIVE', severity: 'info', title: 'Info' },
      { code: 'NO_DNS', severity: 'critical', title: 'Critical' },
      { code: 'SUSPICIOUS_DNS', severity: 'warning', title: 'Warning' }
    ] as ActionItem[];

    expect(primaryAction(actions)?.code).toBe('NO_DNS');
  });

  it('returns no primary action for empty findings', () => {
    expect(primaryAction([])).toBeUndefined();
  });
});
