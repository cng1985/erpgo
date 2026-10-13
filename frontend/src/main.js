import { createApp } from 'vue'
import {
  ElButton, ElCheckbox, ElDialog, ElEmpty, ElForm, ElFormItem, ElIcon, ElInput,
  ElInputNumber, ElLoadingDirective, ElOption, ElPagination, ElProgress, ElSelect,
  ElSkeleton, ElTable, ElTableColumn, ElTag,
} from 'element-plus'
import 'element-plus/dist/index.css'
import {
  ArrowDown, ArrowRight, Bell, Box, Calendar, CirclePlus, Delete, Goods,
  House, Menu, Money, Plus, Search, Setting, ShoppingCart, Tickets, User, Warning,
} from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import './styles/base.css'

const app = createApp(App)
for (const component of [
  ElButton, ElCheckbox, ElDialog, ElEmpty, ElForm, ElFormItem, ElIcon, ElInput,
  ElInputNumber, ElOption, ElPagination, ElProgress, ElSelect, ElSkeleton, ElTable,
  ElTableColumn, ElTag,
]) app.component(component.name, component)
app.directive('loading', ElLoadingDirective)
for (const icon of [ArrowDown, ArrowRight, Bell, Box, Calendar, CirclePlus, Delete, Goods, House, Menu, Money, Plus, Search, Setting, ShoppingCart, Tickets, User, Warning]) app.component(icon.name, icon)
app.use(router).mount('#app')
