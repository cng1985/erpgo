<script setup>
import { computed } from 'vue'
const props = defineProps({ status: { type: String, required: true } })
const config = computed(() => ({
  pending: ['待发货', 'warning'], shipped: ['已发货', 'primary'], completed: ['已完成', 'success'], cancelled: ['已取消', 'info'],
}[props.status] || [props.status, 'info']))
</script>
<template><el-tag :type="config[1]" effect="light" round>{{ config[0] }}</el-tag></template>
