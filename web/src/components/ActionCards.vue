<script setup lang="ts">
import { ref } from 'vue';
import { categoryLabel, severityLabel } from '../domain/overview';
import { useI18n } from '../i18n';
import type { ActionItem } from '../types';

const props = defineProps<{
  primary?: ActionItem;
  secondary: ActionItem[];
}>();

const showSecondary = ref(false);
const i18n = useI18n();
</script>

<template>
  <section class="panel action-panel">
    <div class="section-heading">
      <p class="eyebrow">{{ i18n.t('firstActionEyebrow') }}</p>
      <h2>{{ i18n.t('firstActionTitle') }}</h2>
    </div>

    <article v-if="props.primary" class="primary-action">
      <div class="action-meta">
        <span :class="['badge', props.primary.severity]">{{ severityLabel(props.primary.severity, i18n.locale.value) }}</span>
        <span>{{ categoryLabel(props.primary.category, i18n.locale.value) }}</span>
        <span>{{ i18n.t('confidence') }}: {{ props.primary.confidence }}</span>
      </div>
      <h3>{{ props.primary.title }}</h3>
      <p>{{ props.primary.impact }}</p>

      <div class="steps">
        <h4>{{ i18n.t('howToDo') }}</h4>
        <ol>
          <li v-for="step in props.primary.steps" :key="step">{{ step }}</li>
        </ol>
      </div>

      <div class="verify-box">
        <span>{{ i18n.t('verify') }}</span>
        <p>{{ props.primary.verify }}</p>
      </div>
    </article>

    <div v-else class="empty-state">
      <h3>{{ i18n.t('noActionTitle') }}</h3>
      <p>{{ i18n.t('noActionBody') }}</p>
    </div>

    <button
      v-if="props.secondary.length > 0"
      class="text-button"
      type="button"
      @click="showSecondary = !showSecondary"
    >
      {{ showSecondary ? i18n.t('collapseOtherActions') : i18n.t('showOtherActions', { count: props.secondary.length }) }}
    </button>

    <div v-if="showSecondary" class="secondary-actions">
      <article v-for="action in props.secondary" :key="action.code" class="mini-action">
        <div>
          <strong>{{ action.title }}</strong>
          <p>{{ action.impact }}</p>
        </div>
        <span>{{ categoryLabel(action.category, i18n.locale.value) }}</span>
      </article>
    </div>
  </section>
</template>
