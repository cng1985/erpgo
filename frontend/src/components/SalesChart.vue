<script setup>
import { computed } from 'vue'
import { money } from '../api'
const props = defineProps({ items: { type: Array, default: () => [] } })
const width = 700
const height = 235
const values = computed(() => props.items.map((item) => item.amount || 0))
const max = computed(() => Math.max(10000, ...values.value) * 1.18)
const points = computed(() => values.value.map((value, index) => ({
  x: 52 + index * ((width - 76) / Math.max(1, values.value.length - 1)),
  y: height - 32 - value / max.value * (height - 58),
})))
const line = computed(() => points.value.map((point, i) => `${i ? 'L' : 'M'} ${point.x} ${point.y}`).join(' '))
const area = computed(() => points.value.length ? `${line.value} L ${points.value.at(-1).x} ${height - 32} L ${points.value[0].x} ${height - 32} Z` : '')
const labels = computed(() => props.items.map((item) => item.date?.slice(5).replace('-', '/')))
const ticks = computed(() => [0, .25, .5, .75, 1].map((ratio) => ({ y: height - 32 - ratio * (height - 58), label: Math.round(max.value * ratio / 10000) / 10 })))
</script>
<template>
  <div class="chart-wrap">
    <svg class="sales-chart" :viewBox="`0 0 ${width} ${height}`" role="img" aria-label="近七天销售趋势折线图">
      <defs><linearGradient id="chart-fill" x1="0" x2="0" y1="0" y2="1"><stop offset="0%" stop-color="#3d5bd9" stop-opacity=".17"/><stop offset="100%" stop-color="#3d5bd9" stop-opacity=".01"/></linearGradient></defs>
      <g v-for="tick in ticks" :key="tick.y"><line x1="52" :y1="tick.y" :x2="width - 16" :y2="tick.y" stroke="#e9edf4" stroke-dasharray="3 4"/><text x="42" :y="tick.y + 4" text-anchor="end" class="chart-label">{{ tick.label }}k</text></g>
      <path v-if="area" :d="area" fill="url(#chart-fill)"/><path v-if="line" :d="line" fill="none" stroke="#3d5bd9" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <g v-for="(point, i) in points" :key="i"><circle :cx="point.x" :cy="point.y" r="4" fill="#fff" stroke="#3d5bd9" stroke-width="2.5"/><text :x="point.x" :y="height - 8" text-anchor="middle" class="chart-label">{{ labels[i] }}</text><title>{{ labels[i] }} · {{ money(values[i]) }}</title></g>
    </svg>
    <div class="chart-unit">金额（千元）</div>
  </div>
</template>
