export interface IPLocation {
  country: string;
  city: string;
  org: string;
}

type Fetcher = (url: string) => Promise<{ json: () => Promise<unknown> }>;

interface IPLocationPayload {
  country?: string;
  city?: string;
  org?: string;
}

export function isGeoLookupCandidate(ip: string): boolean {
  const parts = ip.split('.').map((part) => Number(part));
  if (parts.length !== 4 || parts.some((part) => !Number.isInteger(part) || part < 0 || part > 255)) {
    return false;
  }
  const [a, b] = parts;
  if (a === 0 || a === 10 || a === 127) {
    return false;
  }
  if (a === 169 && b === 254) {
    return false;
  }
  if (a === 172 && b >= 16 && b <= 31) {
    return false;
  }
  if (a === 192 && b === 168) {
    return false;
  }
  if (a === 198 && (b === 18 || b === 19)) {
    return false;
  }
  return true;
}

export function formatIPLocation(location?: IPLocation): string {
  if (!location) {
    return '';
  }
  const place = [location.country, location.city].filter(Boolean).join(' ');
  if (!location.org) {
    return place;
  }
  return place ? `${place} · ${location.org}` : location.org;
}

export async function lookupIPLocation(ip: string, fetcher: Fetcher = fetch): Promise<IPLocation | undefined> {
  if (!isGeoLookupCandidate(ip)) {
    return undefined;
  }
  try {
    const response = await fetcher(`/api/ip-location?ip=${encodeURIComponent(ip)}`);
    const payload = (await response.json()) as IPLocationPayload;
    return {
      country: payload.country ?? '',
      city: payload.city ?? '',
      org: payload.org ?? ''
    };
  } catch {
    return undefined;
  }
}
