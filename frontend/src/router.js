import { createRouter, createWebHistory } from 'vue-router'
import { Box, Goods, House, ShoppingCart, Tickets, User } from '@element-plus/icons-vue'
import AppLayout from './components/AppLayout.vue'

export const modules = [
  { path: '/overview', name: 'overview', label: '经营概览', icon: House, component: () => import('./views/DashboardView.vue') },
  { path: '/orders', name: 'orders', label: '销售订单', icon: Tickets, component: () => import('./views/OrdersView.vue') },
  { path: '/products', name: 'products', label: '商品管理', icon: Goods, component: () => import('./views/ProductsView.vue') },
  { path: '/inventory', name: 'inventory', label: '库存管理', icon: Box, component: () => import('./views/InventoryView.vue') },
  { path: '/customers', name: 'customers', label: '客户管理', icon: User, component: () => import('./views/CustomersView.vue') },
  { path: '/purchases', name: 'purchases', label: '采购管理', icon: ShoppingCart, component: () => import('./views/PurchasesView.vue') },
]

export default createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/', component: AppLayout, redirect: '/overview', children: modules.map(({ path, name, component }) => ({ path: path.slice(1), name, component })),
  }, { path: '/:pathMatch(.*)*', redirect: '/overview' }],
})
