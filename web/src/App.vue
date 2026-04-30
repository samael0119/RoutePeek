<script setup lang="ts">
import { computed, ref } from 'vue';
import ActionCards from './components/ActionCards.vue';
import ExplanationDrawer from './components/ExplanationDrawer.vue';
import HealthSummary from './components/HealthSummary.vue';
import NetworkDetails from './components/NetworkDetails.vue';
import TopologyMap from './components/TopologyMap.vue';
import TracePanel from './components/TracePanel.vue';
import { useOverview } from './composables/useOverview';
import { primaryAction } from './domain/overview';
import { useI18n } from './i18n';

const i18n = useI18n();
const overview = useOverview(i18n.locale);
const activeView = ref<'overview' | 'details' | 'trace'>('overview');
const drawerOpen = ref(false);
const drawerTopic = ref('interface');

const firstAction = computed(() => primaryAction(overview.actions.value));
const secondaryActions = computed(() =>
  overview.actions.value.filter((action) => action.code !== firstAction.value?.code)
);

function openTopic(topic: string) {
  drawerTopic.value = topic;
  drawerOpen.value = true;
}
</script>

<template>
  <div class="shell">
    <header class="topbar">
      <div>
        <p class="eyebrow">{{ i18n.t('consoleLabel') }}</p>
        <h1>{{ i18n.t('appTitle') }}</h1>
      </div>
      <div class="topbar-actions">
        <button class="ghost-button" type="button" @click="i18n.toggleLocale">
          {{ i18n.t('language') }}
        </button>
        <button class="primary-button" type="button" :disabled="overview.loading.value" @click="overview.refresh">
          {{ i18n.t('refresh') }}
        </button>
      </div>
    </header>

    <nav class="tabs" aria-label="RoutePeek views">
      <button
        type="button"
        :class="{ active: activeView === 'overview' }"
        @click="activeView = 'overview'"
      >
        {{ i18n.t('overview') }}
      </button>
      <button
        type="button"
        :class="{ active: activeView === 'details' }"
        @click="activeView = 'details'"
      >
        {{ i18n.t('details') }}
      </button>
      <button type="button" :class="{ active: activeView === 'trace' }" @click="activeView = 'trace'">
        {{ i18n.t('trace') }}
      </button>
    </nav>

    <main>
      <section v-if="overview.error.value" class="error-panel">
        <strong>{{ i18n.t('overviewLoadErrorTitle') }}</strong>
        <span>{{ overview.error.value }}</span>
      </section>

      <template v-if="activeView === 'overview'">
        <HealthSummary
          v-if="overview.health.value"
          :health="overview.health.value"
          :loading="overview.loading.value"
        />
        <section v-else class="loading-panel">{{ i18n.t('overviewLoading') }}</section>

        <div class="overview-grid">
          <ActionCards :primary="firstAction" :secondary="secondaryActions" />
          <TopologyMap v-if="overview.topology.value" :topology="overview.topology.value" />
        </div>

        <section v-if="overview.diagnosis.value" class="findings-section">
          <div class="section-heading">
            <p class="eyebrow">{{ i18n.t('otherFindingsEyebrow') }}</p>
            <h2>{{ i18n.t('otherFindingsTitle') }}</h2>
          </div>
          <details
            v-for="finding in overview.diagnosis.value.findings"
            :key="finding.code"
            class="finding-row"
          >
            <summary>
              <span :class="['severity-dot', finding.severity]" />
              <span>{{ finding.title }}</span>
              <code>{{ finding.code }}</code>
            </summary>
            <p>{{ finding.message }}</p>
            <p>{{ finding.suggestion }}</p>
          </details>
          <p v-if="overview.diagnosis.value.findings.length === 0" class="empty-copy">
            {{ i18n.t('noOtherFindings') }}
          </p>
        </section>
      </template>

      <NetworkDetails
        v-else-if="activeView === 'details' && overview.snapshot.value"
        :snapshot="overview.snapshot.value"
        @explain="openTopic"
      />

      <TracePanel v-else-if="activeView === 'trace'" />
    </main>

    <ExplanationDrawer v-model:open="drawerOpen" :topic="drawerTopic" />
  </div>
</template>
