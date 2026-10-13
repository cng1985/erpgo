import { ElMessage } from 'element-plus'

const API_BASE = import.meta.env.VITE_API_BASE || ''

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}/api${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options.headers },
  })
  const data = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(data.error || `请求失败 (${response.status})`)
  return data
}

export const api = {
  dashboard: () => request('/dashboard'),
  customers: (query = '') => request(`/customers${query}`),
  createCustomer: (body) => request('/customers', { method: 'POST', body: JSON.stringify(body) }),
  products: (query = '') => request(`/products${query}`),
  createProduct: (body) => request('/products', { method: 'POST', body: JSON.stringify(body) }),
  orders: (query = '') => request(`/orders${query}`),
  createOrder: (body) => request('/orders', { method: 'POST', body: JSON.stringify(body) }),
  updateOrderStatus: (id, status) => request(`/orders/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  inventory: (query = '') => request(`/inventory${query}`),
  adjustInventory: (body) => request('/inventory/adjust', { method: 'POST', body: JSON.stringify(body) }),
  purchases: () => request('/purchases'),
  receivePurchase: (body) => request('/purchases/receive', { method: 'POST', body: JSON.stringify(body) }),
}

export const money = (cents = 0) => `¥ ${new Intl.NumberFormat('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(cents / 100)}`
export const dateTime = (value) => value ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'short', timeStyle: 'short', timeZone: 'Asia/Shanghai' }).format(new Date(value)) : '—'

export function useRequestError() {
  return (error) => ElMessage.error(error.message || '操作失败，请重试')
}
