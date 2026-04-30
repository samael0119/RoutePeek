<script setup lang="ts">
import { computed } from 'vue';
import { useConnectivity } from '../composables/useConnectivity';
import { useI18n } from '../i18n';
import type { ConnectivityProbeResult } from '../types';

const i18n = useI18n();
const connectivity = useConnectivity(i18n.locale);

const copyLabel = computed(() => {
  if (connectivity.copyState.value === 'copying') {
    return i18n.t('reportCopying');
  }
  if (connectivity.copyState.value === 'success') {
    return i18n.t('reportCopied');
  }
  if (connectivity.copyState.value === 'failed') {
    return i18n.t('reportCopyFailed');
  }
  return i18n.t('reportCopy');
});

function statusLabel(result: ConnectivityProbeResult): string {
  if (!result?.status) {
    return i18n.t('connectivityUnknown');
  }
  const labels: Record<string, string> = {
    ok: i18n.t('connectivityOk'),
    warning: i18n.t('connectivityWarning'),
    danger: i18n.t('connectivityDanger'),
    info: i18n.t('connectivityInfo')
  };
  return labels[result.status] ?? result.status;
}
</script>

<template>
  <section class="panel connectivity-panel">
    <div class="matrix-heading">
      <div>
        <p class="eyebrow">{{ i18n.t('connectivityEyebrow') }}</p>
        <h2>{{ i18n.t('connectivityTitle') }}</h2>
        <p class="plain-copy">
          {{ connectivity.report.value?.summary || i18n.t('connectivityLoading') }}
        </p>
      </div>
      <button class="primary-button" type="button" :disabled="connectivity.copyState.value === 'copying'" @click="connectivity.copyReport">
        {{ copyLabel }}
      </button>
    </div>

    <p v-if="connectivity.error.value" class="form-error">{{ connectivity.error.value }}</p>
    <div v-if="connectivity.loading.value && connectivity.rows.value.length === 0" class="loading-panel compact">
      {{ i18n.t('connectivityLoading') }}
    </div>

    <div v-else class="table-wrap matrix-wrap">
      <table class="matrix-table">
        <thead>
          <tr>
            <th>{{ i18n.t('connectivityTarget') }}</th>
            <th>DNS</th>
            <th>TCP</th>
            <th>HTTP</th>
            <th>{{ i18n.t('connectivityPath') }}</th>
            <th>{{ i18n.t('connectivityConclusion') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in connectivity.rows.value" :key="`${row.target.kind}-${row.target.address}`">
            <td>
              <strong>{{ row.target.name }}</strong>
              <small>{{ row.target.address }}</small>
            </td>
            <td><span :class="['badge', row.dns.status]">{{ statusLabel(row.dns) }}</span></td>
            <td><span :class="['badge', row.tcp.status]">{{ statusLabel(row.tcp) }}</span></td>
            <td><span :class="['badge', row.http.status]">{{ statusLabel(row.http) }}</span></td>
            <td><span :class="['badge', row.path.status]">{{ statusLabel(row.path) }}</span></td>
            <td class="conclusion-cell">{{ row.conclusion }}</td>
          </tr>
          <tr v-if="connectivity.rows.value.length === 0">
            <td colspan="6">{{ i18n.t('connectivityEmpty') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <textarea
      v-if="connectivity.fallbackText.value"
      class="fallback-report"
      :aria-label="i18n.t('reportFallbackLabel')"
      readonly
      :value="connectivity.fallbackText.value"
    />
  </section>
</template>
