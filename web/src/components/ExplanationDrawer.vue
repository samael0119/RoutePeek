<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '../i18n';

const props = defineProps<{
  open: boolean;
  topic: string;
}>();

const emit = defineEmits<{
  'update:open': [open: boolean];
}>();

const i18n = useI18n();

const body = computed(() => {
  const copy: Record<string, ReturnType<typeof i18n.t>> = {
    interface: i18n.t('drawerInterfaceBody'),
    DNS: i18n.t('drawerDnsBody'),
    route: i18n.t('drawerRouteBody'),
    proxy_vpn: i18n.t('drawerProxyVpnBody')
  };
  return copy[props.topic] ?? i18n.t('drawerDefaultBody');
});

const title = computed(() => {
  const titles: Record<string, ReturnType<typeof i18n.t>> = {
    interface: i18n.t('topicInterface'),
    DNS: i18n.t('topicDns'),
    route: i18n.t('topicRoute'),
    proxy_vpn: i18n.t('topicProxyVpn')
  };
  return titles[props.topic] ?? props.topic;
});
</script>

<template>
  <div v-if="props.open" class="drawer-backdrop" @click.self="emit('update:open', false)">
    <aside class="drawer">
      <button class="drawer-close" type="button" @click="emit('update:open', false)">×</button>
      <p class="eyebrow">{{ i18n.t('drawerTerm') }}</p>
      <h2>{{ title }}</h2>
      <p>{{ body }}</p>
    </aside>
  </div>
</template>
