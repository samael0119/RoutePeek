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
  return readJSON<OverviewResponse>(response, locale);
}

export async function fetchTrace(target: string, locale: Locale = 'zh'): Promise<TraceHop[]> {
  const response = await fetch(`/api/trace?target=${encodeURIComponent(target)}`);
  return readJSON<TraceHop[]>(response, locale);
}

export async function fetchConnectivity(locale: Locale = 'zh'): Promise<ConnectivityReport> {
  const response = await fetch(`/api/connectivity?lang=${encodeURIComponent(locale)}`);
  return readJSON<ConnectivityReport>(response, locale);
}

export async function fetchReportMarkdown(locale: Locale = 'zh'): Promise<string> {
  const response = await fetch(`/api/report?format=markdown&lang=${encodeURIComponent(locale)}`);
  if (!response.ok) {
    const payload = (await response.json()) as APIErrorPayload;
    throw new Error(payload.error?.message || tFor(locale, 'requestFailed', { status: response.status }));
  }
  return response.text();
}
