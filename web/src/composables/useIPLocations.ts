import { ref, watch, type Ref } from 'vue';
import { formatIPLocation, isGeoLookupCandidate, lookupIPLocation } from '../domain/ipLocation';
import { type Locale, tFor } from '../i18n';
import type { TraceHop } from '../types';

type LocationState = 'loading' | 'done';

export function useIPLocations(hops: Ref<TraceHop[]>, locale: Ref<Locale>) {
  const labels = ref<Record<string, string>>({});
  const states = ref<Record<string, LocationState>>({});
  const requested = new Set<string>();

  watch(
    hops,
    (nextHops) => {
      for (const hop of nextHops) {
        const ip = hop.address;
        if (!isGeoLookupCandidate(ip) || requested.has(ip)) {
          continue;
        }
        requested.add(ip);
        states.value = { ...states.value, [ip]: 'loading' };
        void lookupIPLocation(ip).then((location) => {
          labels.value = { ...labels.value, [ip]: formatIPLocation(location) };
          states.value = { ...states.value, [ip]: 'done' };
        });
      }
    },
    { immediate: true }
  );

  function labelFor(hop: TraceHop): string {
    const ip = hop.address;
    if (!ip || ip === '*') {
      return '';
    }
    if (!isGeoLookupCandidate(ip)) {
      return tFor(locale.value, 'traceLocationLocal');
    }
    if (states.value[ip] === 'loading') {
      return tFor(locale.value, 'traceLocationLoading');
    }
    return labels.value[ip] || tFor(locale.value, 'traceLocationUnavailable');
  }

  return { labelFor };
}
