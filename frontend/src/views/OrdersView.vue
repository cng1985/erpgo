<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { api, money, dateTime, useRequestError } from '../api'
import PageHeading from '../components/PageHeading.vue'
import StatusTag from '../components/StatusTag.vue'

const route = useRoute(); const router = useRouter(); const onError = useRequestError()
const rows = ref([]); const customers = ref([]); const products = ref([]); const total = ref(0); const query = ref(''); const status = ref(''); const loading = ref(false); const dialog = ref(false); const saving = ref(false)
const page = ref(1); const pageSize = 20
const form = ref({ customer_id: '', note: '', items: [{ product_id: '', quantity: 1 }] })
const totalCents = computed(() => form.value.items.reduce((sum, item) => sum + (products.value.find((p) => p.id === item.product_id)?.price_cents || 0) * (Number(item.quantity) || 0), 0))
async function load() { loading.value = true; try { const params = new URLSearchParams({ page: String(page.value), page_size: String(pageSize) }); if (query.value) params.set('q', query.value); if (status.value) params.set('status', status.value); const data = await api.orders(`?${params}`); rows.value = data.items; total.value = data.total } catch (error) { onError(error) } finally { loading.value = false } }
async function loadOptions() { try { const [c, p] = await Promise.all([api.customers('?page_size=100'), api.products('?page_size=100')]); customers.value = c.items; products.value = p.items } catch (error) { onError(error) } }
function openCreate() { form.value = { customer_id: '', note: '', items: [{ product_id: '', quantity: 1 }] }; dialog.value = true }
function addLine() { form.value.items.push({ product_id: '', quantity: 1 }) }
async function submit() { if (!form.value.customer_id || form.value.items.some((item) => !item.product_id || item.quantity < 1)) { ElMessage.warning('请选择客户和商品，并填写有效数量'); return } saving.value = true; try { await api.createOrder(form.value); ElMessage.success('订单已创建，库存已预占'); dialog.value = false; await load() } catch (error) { onError(error) } finally { saving.value = false } }
async function changeStatus(order, next, label) { try { await ElMessageBox.confirm(`确认将订单 ${order.number} 标记为“${label}”吗？`, '订单状态变更', { type: 'warning' }); await api.updateOrderStatus(order.id, next); ElMessage.success(`订单已更新为${label}`); await load() } catch (error) { if (error !== 'cancel' && error !== 'close') onError(error) } }
watch(() => route.query.status, (value) => { status.value = value || ''; load() }, { immediate: true })
watch(() => route.query.create, (value) => { if (value === '1') { openCreate(); router.replace({ name: 'orders' }) } }, { immediate: true })
onMounted(loadOptions)
</script>

<template>
  <div>
    <PageHeading title="销售订单" subtitle="管理订单、发货和回款流程"><template #default><el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon>新建订单</el-button></template></PageHeading>
    <section class="surface module-surface">
      <div class="filter-bar"><el-input v-model="query" placeholder="搜索订单编号或客户" :prefix-icon="Search" clearable @keyup.enter="page = 1; load()" @clear="page = 1; load()"/><el-select v-model="status" placeholder="全部状态" clearable @change="page = 1; load()"><el-option label="待发货" value="pending"/><el-option label="已发货" value="shipped"/><el-option label="已完成" value="completed"/><el-option label="已取消" value="cancelled"/></el-select><el-button @click="page = 1; load()">筛选</el-button></div>
      <div class="table-scroll"><el-table :data="rows" v-loading="loading" stripe style="width:100%" empty-text="暂无销售订单">
        <el-table-column prop="number" label="订单编号" min-width="180"/><el-table-column label="客户" min-width="190"><template #default="s">{{ s.row.customer?.name }}</template></el-table-column><el-table-column label="商品数量" width="100"><template #default="s">{{ s.row.items?.reduce((n, i) => n + i.quantity, 0) }}</template></el-table-column><el-table-column label="订单金额" width="145"><template #default="s">{{ money(s.row.total_cents) }}</template></el-table-column><el-table-column label="状态" width="115"><template #default="s"><StatusTag :status="s.row.status"/></template></el-table-column><el-table-column label="创建时间" min-width="160"><template #default="s">{{ dateTime(s.row.created_at) }}</template></el-table-column><el-table-column label="操作" min-width="180" fixed="right"><template #default="s"><div class="row-actions"><el-button v-if="s.row.status === 'pending'" link type="primary" @click="changeStatus(s.row,'shipped','已发货')">发货</el-button><el-button v-if="s.row.status === 'pending'" link type="danger" @click="changeStatus(s.row,'cancelled','已取消')">取消</el-button><el-button v-if="s.row.status === 'shipped'" link type="success" @click="changeStatus(s.row,'completed','已完成')">确认回款</el-button><span v-if="['completed','cancelled'].includes(s.row.status)" class="muted">—</span></div></template></el-table-column>
      </el-table></div>
      <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load"/>
    </section>

    <el-dialog v-model="dialog" title="新建销售订单" width="640px" destroy-on-close>
      <el-form label-position="top"><el-form-item label="客户"><el-select v-model="form.customer_id" placeholder="选择客户" filterable><el-option v-for="item in customers" :key="item.id" :label="item.name" :value="item.id"/></el-select></el-form-item>
        <div class="line-heading"><strong>订单商品</strong><el-button link type="primary" @click="addLine"><el-icon><Plus /></el-icon>添加商品</el-button></div>
        <div v-for="(line, index) in form.items" :key="index" class="order-line"><el-select v-model="line.product_id" placeholder="选择商品" filterable><el-option v-for="product in products" :key="product.id" :disabled="product.stock < 1" :label="`${product.name}（库存 ${product.stock}）`" :value="product.id"/></el-select><el-input-number v-model="line.quantity" :min="1" :max="100000"/><el-button :disabled="form.items.length === 1" text type="danger" @click="form.items.splice(index,1)"><el-icon><Delete /></el-icon></el-button></div>
        <el-form-item label="备注"><el-input v-model="form.note" type="textarea" :rows="2" maxlength="500" show-word-limit/></el-form-item>
      </el-form>
      <div class="dialog-total"><span>订单总额</span><strong>{{ money(totalCents) }}</strong></div>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="submit">创建订单</el-button></template>
    </el-dialog>
  </div>
</template>
