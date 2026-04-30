import { describe, expect, it } from 'vitest';
import { formatIPLocation, isGeoLookupCandidate, lookupIPLocation } from './ipLocation';

describe('IP location helpers', () => {
  it('skips local, private, timeout, and proxy fake-ip addresses', () => {
    expect(isGeoLookupCandidate('')).toBe(false);
    expect(isGeoLookupCandidate('*')).toBe(false);
    expect(isGeoLookupCandidate('127.0.0.1')).toBe(false);
    expect(isGeoLookupCandidate('10.0.0.1')).toBe(false);
    expect(isGeoLookupCandidate('172.20.0.1')).toBe(false);
    expect(isGeoLookupCandidate('192.168.1.1')).toBe(false);
    expect(isGeoLookupCandidate('198.18.0.29')).toBe(false);
  });

  it('allows public IPv4 addresses', () => {
    expect(isGeoLookupCandidate('8.8.8.8')).toBe(true);
    expect(isGeoLookupCandidate('39.156.66.10')).toBe(true);
  });

  it('formats location labels with organization when available', () => {
    expect(formatIPLocation({ country: 'China', city: 'Beijing', org: 'Baidu' })).toBe('China Beijing · Baidu');
    expect(formatIPLocation({ country: 'United States', city: '', org: '' })).toBe('United States');
  });

  it('looks up public IP locations through the provided fetcher', async () => {
    const calls: string[] = [];
    const fetcher = async (url: string) => {
      calls.push(url);
      return {
        json: async () => ({ country: 'United States', city: 'Mountain View', org: 'Google LLC' })
      };
    };

    const location = await lookupIPLocation('8.8.8.8', fetcher);

    expect(calls[0]).toBe('/api/ip-location?ip=8.8.8.8');
    expect(location).toEqual({ country: 'United States', city: 'Mountain View', org: 'Google LLC' });
  });
});
