<script setup lang="ts">
import { useIPLocations } from '../composables/useIPLocations';
import { useTrace } from '../composables/useTrace';
import { useI18n } from '../i18n';

const trace = useTrace();
const i18n = useI18n();
const locations = useIPLocations(trace.hops, i18n.locale);
const presets = ['8.8.8.8', '1.1.1.1', 'baidu.com', 'google.com'];
</script>

<template>
  <section class="trace-layout">
    <div class="panel trace-control">
      <div class="section-heading">
        <p class="eyebrow">{{ i18n.t('traceEyebrow') }}</p>
        <h2>{{ trace.explanation.value.panelTitle }}</h2>
      </div>

      <form class="trace-form" @submit.prevent="trace.runTrace()">
        <input
          v-model="trace.target.value"
          type="text"
          autocomplete="off"
          spellcheck="false"
          :aria-label="i18n.t('traceTargetLabel')"
        />
        <button class="primary-button" type="submit" :disabled="trace.loading.value">
          {{ trace.loading.value ? i18n.t('traceRunning') : i18n.t('traceStart') }}
        </button>
      </form>

      <div class="preset-row">
        <button
          v-for="preset in presets"
          :key="preset"
          type="button"
          class="ghost-button"
          @click="trace.runTrace(preset)"
        >
          {{ preset }}
        </button>
      </div>

      <div :class="['trace-explanation', trace.explanation.value.state]">
        <strong>{{ trace.explanation.value.title }}</strong>
        <p>{{ trace.explanation.value.message }}</p>
      </div>
      <p v-if="trace.error.value" class="form-error">{{ trace.error.value }}</p>
    </div>

    <div class="panel hops-panel">
      <table>
        <thead>
          <tr>
            <th>{{ i18n.t('traceHop') }}</th>
            <th>{{ i18n.t('traceAddress') }}</th>
            <th>{{ i18n.t('traceLocation') }}</th>
            <th>{{ i18n.t('traceRtt1') }}</th>
            <th>{{ i18n.t('traceRtt2') }}</th>
            <th>{{ i18n.t('traceRtt3') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="hop in trace.hops.value" :key="hop.hop">
            <td>{{ hop.hop }}</td>
            <td>{{ hop.hostname || hop.address || '*' }}</td>
            <td>{{ locations.labelFor(hop) }}</td>
            <td>{{ hop.rtt1 }}</td>
            <td>{{ hop.rtt2 }}</td>
            <td>{{ hop.rtt3 }}</td>
          </tr>
          <tr v-if="trace.hops.value.length === 0">
            <td colspan="6">{{ i18n.t('traceNoResults') }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
