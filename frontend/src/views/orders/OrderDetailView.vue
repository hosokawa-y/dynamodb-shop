<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useOrderStore } from '@/stores/order'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const orderStore = useOrderStore()
const authStore = useAuthStore()

const orderId = route.params.id as string

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: `/orders/${orderId}` } })
    return
  }

  orderStore.clearCurrentOrder()
  await orderStore.fetchOrderById(orderId)
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
    hour: '2-digit',
    minute: '2-digit',
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

function goBack() {
  router.push('/orders')
}
</script>

<template>
  <div class="order-detail-view">
    <button class="btn-back" @click="goBack">← Back to Orders</button>

    <div v-if="orderStore.loading" class="loading">Loading order details...</div>

    <div v-else-if="orderStore.error" class="error">
      {{ orderStore.error }}
    </div>

    <div v-else-if="orderStore.currentOrder" class="order-detail">
      <div class="order-header">
        <div>
          <h1>Order Details</h1>
          <p class="order-id">Order #{{ orderStore.currentOrder.id }}</p>
        </div>
        <div :class="['order-status', getStatusClass(orderStore.currentOrder.status)]">
          {{ orderStore.currentOrder.status }}
        </div>
      </div>

      <div class="order-info-card">
        <h2>Order Information</h2>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Order Date</span>
            <span class="value">{{ formatDate(orderStore.currentOrder.createdAt) }}</span>
          </div>
          <div class="info-item">
            <span class="label">Items</span>
            <span class="value">{{ orderStore.currentOrder.itemCount }} items</span>
          </div>
          <div class="info-item">
            <span class="label">Total Amount</span>
            <span class="value total">{{ formatPrice(orderStore.currentOrder.totalAmount) }}</span>
          </div>
        </div>
      </div>

      <div class="order-items-card">
        <h2>Order Items</h2>
        <div class="items-list">
          <div
            v-for="item in orderStore.currentOrder.items"
            :key="item.productId"
            class="order-item"
          >
            <div class="item-info">
              <span class="item-name">{{ item.productName }}</span>
              <span class="item-meta">
                {{ formatPrice(item.price) }} × {{ item.quantity }}
              </span>
            </div>
            <div class="item-subtotal">
              {{ formatPrice(item.subtotal) }}
            </div>
          </div>
        </div>

        <div class="items-total">
          <span>Total</span>
          <span>{{ formatPrice(orderStore.currentOrder.totalAmount) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.order-detail-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.btn-back {
  background: none;
  border: none;
  color: #4a90d9;
  cursor: pointer;
  padding: 0;
  margin-bottom: 1.5rem;
  font-size: 0.95rem;
}

.btn-back:hover {
  text-decoration: underline;
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

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
}

h1 {
  margin: 0 0 0.25rem 0;
  color: #333;
}

.order-id {
  color: #666;
  font-size: 0.9rem;
  margin: 0;
}

.order-status {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  font-size: 0.9rem;
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

.order-info-card,
.order-items-card {
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

h2 {
  font-size: 1.1rem;
  margin: 0 0 1rem 0;
  color: #333;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media (max-width: 600px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-item .label {
  color: #666;
  font-size: 0.85rem;
}

.info-item .value {
  color: #333;
  font-weight: 500;
}

.info-item .value.total {
  font-size: 1.1rem;
  color: #28a745;
}

.items-list {
  display: flex;
  flex-direction: column;
}

.order-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 0;
  border-bottom: 1px solid #f0f0f0;
}

.order-item:last-child {
  border-bottom: none;
}

.item-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.item-name {
  font-weight: 500;
  color: #333;
}

.item-meta {
  color: #666;
  font-size: 0.9rem;
}

.item-subtotal {
  font-weight: 500;
  color: #333;
}

.items-total {
  display: flex;
  justify-content: space-between;
  padding-top: 1rem;
  margin-top: 0.5rem;
  border-top: 2px solid #eee;
  font-size: 1.1rem;
  font-weight: 600;
  color: #333;
}
</style>
