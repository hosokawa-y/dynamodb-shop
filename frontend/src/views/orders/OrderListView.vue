<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { useAuthStore } from '@/stores/auth'

const orderStore = useOrderStore()
const authStore = useAuthStore()
const router = useRouter()

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: '/orders' } })
    return
  }
  await orderStore.fetchOrders()
})

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'CONFIRMED':
      return 'status-confirmed'
    case 'SHIPPED':
      return 'status-shipped'
    case 'DELIVERED':
      return 'status-delivered'
    case 'CANCELLED':
      return 'status-cancelled'
    default:
      return 'status-pending'
  }
}

function viewOrderDetail(orderId: string) {
  router.push({ name: 'order-detail', params: { id: orderId } })
}

function goToProducts() {
  router.push('/products')
}
</script>

<template>
  <div class="orders-view">
    <h1>Order History</h1>

    <div v-if="orderStore.loading" class="loading">Loading orders...</div>

    <div v-else-if="orderStore.error" class="error">
      {{ orderStore.error }}
    </div>

    <div v-else-if="orderStore.orders.length === 0" class="empty">
      <p>No orders yet</p>
      <button class="btn-primary" @click="goToProducts">Start Shopping</button>
    </div>

    <div v-else class="orders-list">
      <div
        v-for="order in orderStore.orders"
        :key="order.id"
        class="order-card"
        @click="viewOrderDetail(order.id)"
      >
        <div class="order-header">
          <div class="order-date">{{ formatDate(order.createdAt) }}</div>
          <div :class="['order-status', getStatusClass(order.status)]">
            {{ order.status }}
          </div>
        </div>

        <div class="order-body">
          <div class="order-id">Order #{{ order.id.slice(0, 8) }}...</div>
          <div class="order-info">
            <span>{{ order.itemCount }} items</span>
            <span class="order-total">{{ formatPrice(order.totalAmount) }}</span>
          </div>
        </div>

        <div class="order-footer">
          <span class="view-detail">View Details →</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.orders-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

h1 {
  margin-bottom: 2rem;
  color: #333;
}

.loading,
.error {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error {
  color: #c00;
}

.empty {
  text-align: center;
  padding: 3rem;
  background: #f9f9f9;
  border-radius: 8px;
}

.empty p {
  margin-bottom: 1.5rem;
  color: #666;
  font-size: 1.1rem;
}

.btn-primary {
  background: #4a90d9;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-primary:hover {
  background: #3a7bc8;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.order-card {
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 1.25rem;
  cursor: pointer;
  transition: box-shadow 0.2s, border-color 0.2s;
}

.order-card:hover {
  border-color: #4a90d9;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.order-date {
  color: #666;
  font-size: 0.9rem;
}

.order-status {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-confirmed {
  background: #e8f5e9;
  color: #2e7d32;
}

.status-shipped {
  background: #e3f2fd;
  color: #1565c0;
}

.status-delivered {
  background: #f3e5f5;
  color: #7b1fa2;
}

.status-cancelled {
  background: #ffebee;
  color: #c62828;
}

.status-pending {
  background: #fff3e0;
  color: #ef6c00;
}

.order-body {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.order-id {
  font-weight: 500;
  color: #333;
}

.order-info {
  display: flex;
  gap: 1.5rem;
  color: #666;
}

.order-total {
  font-weight: 600;
  color: #333;
}

.order-footer {
  border-top: 1px solid #f0f0f0;
  padding-top: 0.75rem;
}

.view-detail {
  color: #4a90d9;
  font-size: 0.9rem;
}
</style>
