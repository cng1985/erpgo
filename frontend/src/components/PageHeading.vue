<script setup>
defineProps({ title: { type: String, required: true }, subtitle: { type: String, default: '' }, mobileSubtitle: { type: String, default: '' } })
defineEmits(['action'])
</script>
<template>
  <div class="page-heading">
    <div><h1>{{ title }}</h1><p v-if="subtitle" class="page-subtitle-desktop">{{ subtitle }}</p><p v-if="mobileSubtitle" class="page-subtitle-mobile">{{ mobileSubtitle }}</p></div>
    <slot><el-button type="primary" @click="$emit('action')"><el-icon><Plus /></el-icon>新建</el-button></slot>
  </div>
</template>
