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
    router.push({ name: 'login' })
    return
  }

  if (orderId) {
    await orderStore.fetchOrderById(orderId)
  }
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

function goToOrders() {
  router.push('/orders')
}

function goToProducts() {
  router.push('/products')
}
</script>

<template>
  <div class="order-complete-view">
    <div class="success-icon">✓</div>
    <h1>Order Placed Successfully!</h1>
    <p class="thank-you">Thank you for your order.</p>

    <div v-if="orderStore.loading" class="loading">Loading order details...</div>

    <div v-else-if="orderStore.currentOrder" class="order-details">
      <div class="detail-row">
        <span class="label">Order ID</span>
        <span class="value">{{ orderStore.currentOrder.id }}</span>
      </div>
      <div class="detail-row">
        <span class="label">Order Date</span>
        <span class="value">{{ formatDate(orderStore.currentOrder.createdAt) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">Status</span>
        <span class="value status">{{ orderStore.currentOrder.status }}</span>
      </div>
      <div class="detail-row">
        <span class="label">Items</span>
        <span class="value">{{ orderStore.currentOrder.itemCount }} items</span>
      </div>
      <div class="detail-row total">
        <span class="label">Total</span>
        <span class="value">{{ formatPrice(orderStore.currentOrder.totalAmount) }}</span>
      </div>
    </div>

    <div class="actions">
      <button class="btn-orders" @click="goToOrders">View Order History</button>
      <button class="btn-continue" @click="goToProducts">Continue Shopping</button>
    </div>
  </div>
</template>

<style scoped>
.order-complete-view {
  max-width: 500px;
  margin: 0 auto;
  padding: 3rem 1rem;
  text-align: center;
}

.success-icon {
  width: 80px;
  height: 80px;
  background: #28a745;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.5rem;
  margin: 0 auto 1.5rem;
}

h1 {
  color: #333;
  margin-bottom: 0.5rem;
}

.thank-you {
  color: #666;
  margin-bottom: 2rem;
}

.loading {
  color: #666;
  padding: 2rem;
}

.order-details {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  text-align: left;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid #eee;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-row.total {
  border-top: 2px solid #ddd;
  margin-top: 0.5rem;
  padding-top: 1rem;
  font-weight: 600;
}

.label {
  color: #666;
}

.value {
  color: #333;
  font-weight: 500;
}

.value.status {
  background: #e8f5e9;
  color: #2e7d32;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.btn-orders {
  width: 100%;
  padding: 0.875rem;
  background: #4a90d9;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
}

.btn-orders:hover {
  background: #3a7bc8;
}

.btn-continue {
  width: 100%;
  padding: 0.875rem;
  background: white;
  color: #666;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
}

.btn-continue:hover {
  background: #f5f5f5;
}
</style>
