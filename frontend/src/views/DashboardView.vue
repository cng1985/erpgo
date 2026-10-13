<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, money, dateTime, useRequestError } from '../api'
import SalesChart from '../components/SalesChart.vue'
import StatusTag from '../components/StatusTag.vue'
import PageHeading from '../components/PageHeading.vue'

const router = useRouter()
const onError = useRequestError()
const summary = ref(null)
const busy = ref(false)
const today = computed(() => new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', timeZone: 'Asia/Shanghai' }).format(new Date()))
const actions = computed(() => (summary.value?.recent_orders || []).filter((order) => order.status === 'pending').slice(0, 2))
async function load() { busy.value = true; try { summary.value = await api.dashboard() } catch (error) { onError(error) } finally { busy.value = false } }
onMounted(load)
</script>

<template>
  <div class="dashboard-page">
    <PageHeading title="经营概览" subtitle="查看企业运营情况" :mobile-subtitle="today">
      <template #default><div class="heading-actions"><span class="date-control"><el-icon><Calendar /></el-icon>{{ today }}</span><el-button type="primary" @click="router.push({ name: 'orders', query: { create: '1' } })"><el-icon><Plus /></el-icon>新建订单</el-button></div></template>
    </PageHeading>

    <div v-if="busy && !summary" class="loading-state"><el-skeleton :rows="7" animated /></div>
    <template v-else-if="summary">
      <section class="metrics-grid" aria-label="经营数据">
        <button class="metric-card" @click="router.push({ name: 'orders' })"><span class="metric-icon blue"><el-icon><ShoppingCart /></el-icon></span><span class="metric-copy"><span class="metric-label">今日销售额</span><strong>{{ money(summary.today_sales_cents) }}</strong><small>来自今日销售订单</small></span></button>
        <button class="metric-card" @click="router.push({ name: 'orders', query: { status: 'pending' } })"><span class="metric-icon orange"><el-icon><Tickets /></el-icon></span><span class="metric-copy"><span class="metric-label">待发货订单</span><strong>{{ summary.pending_orders }}</strong><small>等待处理</small></span></button>
        <button class="metric-card" @click="router.push({ name: 'inventory', query: { low_stock: '1' } })"><span class="metric-icon red"><el-icon><Warning /></el-icon></span><span class="metric-copy"><span class="metric-label">库存预警商品</span><strong>{{ summary.low_stock_products }}</strong><small>低于安全库存</small></span></button>
        <button class="metric-card" @click="router.push({ name: 'orders', query: { status: 'shipped' } })"><span class="metric-icon green"><el-icon><Money /></el-icon></span><span class="metric-copy"><span class="metric-label">应收账款</span><strong>{{ money(summary.receivable_cents) }}</strong><small>已发货待回款</small></span></button>
      </section>

      <section class="overview-grid">
        <article class="surface chart-panel"><header class="surface-heading"><h2>销售趋势</h2><span>近 7 天</span></header><SalesChart :items="summary.sales_trend" /></article>
        <article class="surface task-panel"><header class="surface-heading"><h2>待处理事项</h2><el-button text type="primary" @click="router.push({ name: 'orders', query: { status: 'pending' } })">更多 <el-icon><ArrowRight /></el-icon></el-button></header>
          <button v-for="order in actions" :key="order.id" class="task-row" @click="router.push({ name: 'orders' })"><div class="task-content"><StatusTag :status="order.status"/><div><strong>{{ order.number }}</strong><span>{{ order.customer?.name }} · {{ order.items?.length || 0 }} 种商品</span></div><b>{{ money(order.total_cents) }}</b></div><el-icon><ArrowRight /></el-icon></button>
          <el-empty v-if="!actions.length" description="当前没有待处理订单" :image-size="72" />
        </article>
      </section>

      <section class="surface recent-panel"><header class="surface-heading"><h2>最近订单</h2><el-button text type="primary" @click="router.push({ name: 'orders' })">全部订单 <el-icon><ArrowRight /></el-icon></el-button></header>
        <div class="table-scroll recent-desktop-table"><el-table :data="summary.recent_orders" stripe style="width: 100%" empty-text="暂无订单">
          <el-table-column prop="number" label="订单编号" min-width="175" />
          <el-table-column label="客户" min-width="190"><template #default="scope">{{ scope.row.customer?.name || '—' }}</template></el-table-column>
          <el-table-column label="商品数量" width="115"><template #default="scope">{{ scope.row.items?.reduce((sum, item) => sum + item.quantity, 0) || 0 }}</template></el-table-column>
          <el-table-column label="金额" width="135"><template #default="scope">{{ money(scope.row.total_cents) }}</template></el-table-column>
          <el-table-column label="状态" width="120"><template #default="scope"><StatusTag :status="scope.row.status" /></template></el-table-column>
          <el-table-column label="日期" min-width="165"><template #default="scope">{{ dateTime(scope.row.created_at) }}</template></el-table-column>
          <el-table-column label="操作" width="72" fixed="right"><template #default><el-button link type="primary" @click="router.push({ name: 'orders' })">查看</el-button></template></el-table-column>
        </el-table></div>
        <div class="mobile-order-table">
          <div class="mobile-order-header"><span>订单编号</span><span>客户</span><span>金额</span><span>状态</span></div>
          <button v-for="order in summary.recent_orders" :key="order.id" class="mobile-order-row" @click="router.push({ name: 'orders' })">
            <span class="mobile-order-number">{{ order.number }}</span><span class="mobile-order-customer" :title="order.customer?.name">{{ order.customer?.name || '—' }}</span><span class="mobile-order-amount">{{ money(order.total_cents) }}</span><StatusTag :status="order.status" />
          </button>
        </div>
      </section>
    </template>
  </div>
</template>
