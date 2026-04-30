import { computed, onMounted, ref, watch, type Ref } from 'vue';
import { fetchConnectivity, fetchReportMarkdown } from '../api';
import { type Locale, tFor } from '../i18n';
import type { ConnectivityReport } from '../types';

export function useConnectivity(locale: Ref<Locale>) {
  const report = ref<ConnectivityReport>();
  const loading = ref(false);
  const error = ref('');
  const copyState = ref<'idle' | 'copying' | 'success' | 'failed'>('idle');
  const fallbackText = ref('');

  async function refresh() {
    loading.value = true;
    error.value = '';
    try {
      report.value = await fetchConnectivity(locale.value);
    } catch (err) {
      error.value = err instanceof Error ? err.message : tFor(locale.value, 'connectivityLoadError');
    } finally {
      loading.value = false;
    }
  }

  async function copyReport() {
    copyState.value = 'copying';
    fallbackText.value = '';
    try {
      const markdown = await fetchReportMarkdown(locale.value);
      if (!navigator.clipboard?.writeText) {
        fallbackText.value = markdown;
        copyState.value = 'failed';
        return;
      }
      try {
        await navigator.clipboard.writeText(markdown);
      } catch {
        fallbackText.value = markdown;
        copyState.value = 'failed';
        return;
      }
      copyState.value = 'success';
      window.setTimeout(() => {
        if (copyState.value === 'success') {
          copyState.value = 'idle';
        }
      }, 1800);
    } catch (err) {
      copyState.value = 'failed';
      fallbackText.value = err instanceof Error ? err.message : tFor(locale.value, 'reportCopyFailed');
    }
  }

  onMounted(refresh);
  watch(locale, refresh);

  return {
    report,
    loading,
    error,
    copyState,
    fallbackText,
    refresh,
    copyReport,
    rows: computed(() => report.value?.targets ?? [])
  };
}
