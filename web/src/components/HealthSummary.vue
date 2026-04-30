<script setup lang="ts">
import type { HealthOverview } from '../types';
import { healthTone } from '../domain/overview';
import { useI18n } from '../i18n';

const props = defineProps<{
  health: HealthOverview;
  loading: boolean;
}>();

const i18n = useI18n();
</script>

<template>
  <section :class="['health-summary', healthTone(props.health)]">
    <div class="status-block">
      <p class="eyebrow">{{ i18n.t('healthEyebrow') }}</p>
      <div class="status-line">
        <span class="pulse" />
        <h2>{{ props.health.label }}</h2>
      </div>
      <p>{{ props.health.summary }}</p>
    </div>
    <div class="risk-block">
      <span>{{ i18n.t('riskLevel') }}</span>
      <strong>{{ props.health.risk_level.toUpperCase() }}</strong>
      <small v-if="props.health.primary_issue">{{ props.health.primary_issue }}</small>
      <small v-else>{{ i18n.t('noPrimaryRisk') }}</small>
    </div>
    <div v-if="props.loading" class="refreshing">{{ i18n.t('refreshing') }}</div>
  </section>
</template>
