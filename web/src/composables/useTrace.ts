import { computed, ref } from 'vue';
import { fetchTrace } from '../api';
import { explainTrace } from '../domain/trace';
import { type Locale, tFor, useI18n } from '../i18n';
import type { TraceHop } from '../types';

export function useTrace(initialTarget = '8.8.8.8') {
  const { locale } = useI18n();
  const target = ref(initialTarget);
  const hops = ref<TraceHop[]>([]);
  const loading = ref(false);
  const error = ref('');

  async function runTrace(nextTarget = target.value) {
    target.value = nextTarget.trim();
    if (!target.value) {
      error.value = tFor(locale.value as Locale, 'traceMissingTarget');
      return;
    }
    loading.value = true;
    error.value = '';
    try {
      hops.value = await fetchTrace(target.value, locale.value as Locale);
    } catch (err) {
      error.value = err instanceof Error ? err.message : tFor(locale.value as Locale, 'traceFailed');
    } finally {
      loading.value = false;
    }
  }

  return {
    target,
    hops,
    loading,
    error,
    explanation: computed(() => explainTrace(hops.value, locale.value as Locale)),
    runTrace
  };
}
