<script setup lang="ts">
import { computed } from 'vue';
import type { NetworkTopology, TopologyNode } from '../types';
import { useI18n } from '../i18n';

const props = defineProps<{
  topology: NetworkTopology;
}>();

const i18n = useI18n();

const positions: Record<string, { x: number; y: number }> = {
  device: { x: 80, y: 150 },
  gateway: { x: 250, y: 150 },
  internet: { x: 430, y: 150 },
  dns: { x: 250, y: 55 },
  vpn: { x: 250, y: 245 },
  proxy: { x: 365, y: 245 },
  vm: { x: 80, y: 245 }
};

const nodes = computed(() =>
  props.topology.nodes.map((node) => ({
    ...node,
    ...(positions[node.id] ?? { x: 250, y: 150 })
  }))
);

function pointFor(id: string) {
  return positions[id] ?? { x: 250, y: 150 };
}

function nodeClass(node: TopologyNode) {
  return ['topology-node', node.status === 'risk' ? 'risk' : 'ok', node.kind];
}
</script>

<template>
  <section class="panel topology-panel">
    <div class="section-heading">
      <p class="eyebrow">{{ i18n.t('topologyEyebrow') }}</p>
      <h2>{{ i18n.t('topologyTitle') }}</h2>
    </div>

    <svg viewBox="0 0 510 310" role="img" :aria-label="i18n.t('topologyAriaLabel')">
      <defs>
        <filter id="softGlow">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      <line
        v-for="link in props.topology.links"
        :key="`${link.from}-${link.to}`"
        :x1="pointFor(link.from).x"
        :y1="pointFor(link.from).y"
        :x2="pointFor(link.to).x"
        :y2="pointFor(link.to).y"
        :class="['topology-link', link.status]"
      />

      <g v-for="node in nodes" :key="node.id" :class="nodeClass(node)" filter="url(#softGlow)">
        <circle :cx="node.x" :cy="node.y" r="30" />
        <text :x="node.x" :y="node.y + 5" text-anchor="middle">{{ node.label }}</text>
        <text :x="node.x" :y="node.y + 48" text-anchor="middle" class="node-description">
          {{ node.description }}
        </text>
      </g>
    </svg>
  </section>
</template>
