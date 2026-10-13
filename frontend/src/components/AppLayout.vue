<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { modules } from '../router'
import { Search, User } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const mobileMenu = ref(false)
const current = computed(() => modules.find((item) => item.name === route.name)?.label || '经营概览')
const bottomModules = computed(() => [
  { ...modules[0], label: '概览' }, { ...modules[1], label: '订单' }, { ...modules[2], label: '商品' }, { ...modules[3], label: '库存' },
  { name: 'more', label: '我的', icon: User },
])
function navigate(name) {
  if (name === 'more') { mobileMenu.value = !mobileMenu.value; return }
  router.push({ name })
  mobileMenu.value = false
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand"><img class="brand-mark" src="/brand-symbol.png" alt=""/><span>北辰 <b>ERP</b></span></div>
      <nav class="side-nav" aria-label="主导航">
        <button v-for="item in modules" :key="item.name" class="nav-item" :class="{ active: route.name === item.name }" @click="navigate(item.name)">
          <el-icon><component :is="item.icon" /></el-icon><span>{{ item.label }}</span>
        </button>
        <button class="nav-item settings-item" @click="$message.info('系统设置即将开放')"><el-icon><Setting /></el-icon><span>系统设置</span></button>
      </nav>
      <div class="sidebar-foot"><span class="online-dot"></span> 本地数据库已连接</div>
    </aside>

    <section class="workspace">
      <div class="mobile-brand"><span>北辰 <b>ERP</b></span><button class="mobile-menu-button" aria-label="打开菜单" @click="mobileMenu = !mobileMenu"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 12h16M4 18h16"/></svg></button></div>
      <header class="topbar">
        <div class="breadcrumb"><span>工作台</span><el-icon><ArrowRight /></el-icon><strong>{{ current }}</strong></div>
        <div class="top-actions">
          <el-input class="global-search" placeholder="搜索订单、客户、商品等…" :prefix-icon="Search" clearable @keyup.enter="navigate('orders')" />
          <button class="icon-button" aria-label="通知" @click="$message.info('暂无新通知')"><el-icon><Bell /></el-icon></button>
          <button class="profile-button"><span class="avatar">张</span><span class="profile-name">张三</span><el-icon><ArrowDown /></el-icon></button>
        </div>
      </header>
      <main class="main-content"><router-view /></main>
    </section>

    <div v-if="mobileMenu" class="mobile-more-menu">
      <div class="mobile-profile"><span class="avatar">张</span><span><strong>张三</strong><small>北辰 ERP 工作台</small></span></div>
      <button v-for="item in modules.slice(4)" :key="item.name" @click="navigate(item.name)"><el-icon><component :is="item.icon" /></el-icon>{{ item.label }}</button>
      <button @click="$message.info('系统设置即将开放')"><el-icon><Setting /></el-icon>系统设置</button>
    </div>
    <nav class="mobile-nav" aria-label="移动端主导航">
      <button v-for="item in bottomModules" :key="item.name" :class="{ active: route.name === item.name || item.name === 'more' && mobileMenu }" @click="navigate(item.name)">
        <el-icon><component :is="item.icon" /></el-icon><span>{{ item.label }}</span>
      </button>
    </nav>
  </div>
</template>
