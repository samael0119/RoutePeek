import { describe, expect, it } from 'vitest';
import { messages, tFor } from './i18n';

describe('frontend i18n dictionary', () => {
  it('keeps zh and en key sets in sync', () => {
    const zhKeys = Object.keys(messages.zh).sort();
    const enKeys = Object.keys(messages.en).sort();

    expect(enKeys).toEqual(zhKeys);
  });

  it('translates shared UI labels by locale', () => {
    expect(tFor('zh', 'traceStart')).toBe('开始');
    expect(tFor('en', 'traceStart')).toBe('Start');
    expect(tFor('en', 'detailsInterfacesTitle')).toBe('Interfaces');
    expect(tFor('en', 'detailsNoRoutes')).toBe('No route entries detected.');
    expect(tFor('zh', 'traceEyebrow')).toBe('路径追踪');
  });
});
