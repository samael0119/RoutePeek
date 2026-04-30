<script setup lang="ts">
import type { NetworkSnapshot } from '../types';
import { useI18n } from '../i18n';

const props = defineProps<{
  snapshot: NetworkSnapshot;
}>();

defineEmits<{
  explain: [topic: string];
}>();

const i18n = useI18n();
</script>

<template>
  <section class="details-layout">
    <article class="panel detail-panel">
      <button class="info-button" type="button" @click="$emit('explain', 'interface')">?</button>
      <p class="eyebrow">{{ i18n.t('detailsInterfacesEyebrow') }}</p>
      <h2>{{ i18n.t('detailsInterfacesTitle') }}</h2>
      <p class="plain-copy">{{ i18n.t('detailsInterfacesCopy') }}</p>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>{{ i18n.t('detailsName') }}</th>
              <th>{{ i18n.t('detailsIpv4') }}</th>
              <th>{{ i18n.t('detailsType') }}</th>
              <th>{{ i18n.t('detailsStatus') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="iface in props.snapshot.interfaces" :key="iface.name">
              <td>{{ iface.name }}</td>
              <td>{{ iface.ip4 || '-' }}</td>
              <td>{{ iface.type || '-' }}</td>
              <td>{{ iface.is_up ? i18n.t('detailsEnabled') : i18n.t('detailsDisabled') }}</td>
            </tr>
            <tr v-if="props.snapshot.interfaces.length === 0">
              <td colspan="4">{{ i18n.t('detailsNoInterfaces') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <article class="panel detail-panel">
      <button class="info-button" type="button" @click="$emit('explain', 'DNS')">?</button>
      <p class="eyebrow">{{ i18n.t('detailsDnsTitle') }}</p>
      <h2>{{ i18n.t('detailsDnsTitle') }}</h2>
      <p class="plain-copy">{{ i18n.t('detailsDnsCopy') }}</p>
      <div class="chips">
        <span v-for="server in props.snapshot.dns.servers" :key="server">{{ server }}</span>
        <span v-if="props.snapshot.dns.servers.length === 0">{{ i18n.t('detailsNotConfigured') }}</span>
      </div>
    </article>

    <article class="panel detail-panel">
      <button class="info-button" type="button" @click="$emit('explain', 'route')">?</button>
      <p class="eyebrow">{{ i18n.t('detailsRoutesEyebrow') }}</p>
      <h2>{{ i18n.t('detailsRoutesTitle') }}</h2>
      <p class="plain-copy">{{ i18n.t('detailsRoutesCopy') }}</p>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>{{ i18n.t('detailsDestination') }}</th>
              <th>{{ i18n.t('detailsGateway') }}</th>
              <th>{{ i18n.t('detailsInterface') }}</th>
              <th>{{ i18n.t('detailsMetric') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="route in props.snapshot.routes.slice(0, 12)" :key="`${route.destination}-${route.interface}-${route.metric}`">
              <td>{{ route.destination }}</td>
              <td>{{ route.gateway || '-' }}</td>
              <td>{{ route.interface || '-' }}</td>
              <td>{{ route.metric }}</td>
            </tr>
            <tr v-if="props.snapshot.routes.length === 0">
              <td colspan="4">{{ i18n.t('detailsNoRoutes') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <article class="panel detail-panel">
      <button class="info-button" type="button" @click="$emit('explain', 'proxy_vpn')">?</button>
      <p class="eyebrow">{{ i18n.t('detailsProxyVpnVmEyebrow') }}</p>
      <h2>{{ i18n.t('detailsProxyVpnVmTitle') }}</h2>
      <p class="plain-copy">{{ i18n.t('detailsProxyVpnVmCopy') }}</p>
      <dl class="kv-list">
        <div>
          <dt>{{ i18n.t('detailsProxy') }}</dt>
          <dd>{{ props.snapshot.proxy.has_proxy ? i18n.t('detailsProxyOn') : i18n.t('detailsProxyOff') }}</dd>
        </div>
        <div>
          <dt>{{ i18n.t('detailsVpn') }}</dt>
          <dd>{{ props.snapshot.vpn?.status === 'connected' ? props.snapshot.vpn.name : i18n.t('detailsVpnDisconnected') }}</dd>
        </div>
        <div>
          <dt>{{ i18n.t('detailsVmNetworks') }}</dt>
          <dd>{{ props.snapshot.vm_networks.length }} {{ i18n.t('detailsCountSuffix') }}</dd>
        </div>
        <div>
          <dt>{{ i18n.t('detailsPublicIp') }}</dt>
          <dd>{{ props.snapshot.public_ip || i18n.t('detailsNotDetected') }}</dd>
        </div>
      </dl>
    </article>
  </section>
</template>
