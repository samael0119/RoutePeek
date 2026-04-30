import { computed, onMounted, ref, watch, type Ref } from 'vue';
import { fetchOverview } from '../api';
import { type Locale, tFor } from '../i18n';
import type { OverviewResponse } from '../types';

export function useOverview(locale: Ref<Locale>) {
  const data = ref<OverviewResponse>();
  const loading = ref(false);
  const error = ref('');

  async function refresh() {
    loading.value = true;
    error.value = '';
    try {
      data.value = await fetchOverview(locale.value);
    } catch (err) {
      error.value = err instanceof Error ? err.message : tFor(locale.value, 'overviewLoadError');
    } finally {
      loading.value = false;
    }
  }

  onMounted(refresh);
  watch(locale, refresh);

  return {
    data,
    loading,
    error,
    refresh,
    snapshot: computed(() => data.value?.snapshot),
    health: computed(() => data.value?.health),
    actions: computed(() => data.value?.actions ?? []),
    topology: computed(() => data.value?.topology),
    diagnosis: computed(() => data.value?.diagnosis)
  };
}
