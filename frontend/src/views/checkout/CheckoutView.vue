<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { useOrderStore } from '@/stores/order'
import { useAuthStore } from '@/stores/auth'

const cartStore = useCartStore()
const orderStore = useOrderStore()
const authStore = useAuthStore()
const router = useRouter()

const submitting = ref(false)
const orderError = ref<string | null>(null)

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: '/checkout' } })
    return
  }
  await cartStore.fetchCart()

  // カートが空の場合はカートページにリダイレクト
  if (cartStore.itemCount === 0) {
    router.push('/cart')
  }
})

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}

async function placeOrder() {
  if (submitting.value) return

  submitting.value = true
  orderError.value = null

  try {
    const order = await orderStore.createOrder()
    // 注文完了後、カートをクリア
    cartStore.clearCart()
    // 注文完了ページに遷移
    router.push({ name: 'order-complete', params: { id: order.id } })
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    orderError.value = err.response?.data?.error || 'Failed to place order'
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.push('/cart')
}
</script>

<template>
  <div class="checkout-view">
    <h1>Order Confirmation</h1>

    <div v-if="cartStore.loading" class="loading">Loading...</div>

    <div v-else-if="cartStore.itemCount === 0" class="empty">
      <p>Your cart is empty</p>
    </div>

    <div v-else class="checkout-content">
      <div class="order-items">
        <h2>Order Items</h2>
        <div class="items-list">
          <div v-for="item in cartStore.items" :key="item.productId" class="order-item">
            <div class="item-info">
              <span class="item-name">{{ item.productName }}</span>
              <span class="item-qty">x {{ item.quantity }}</span>
            </div>
            <div class="item-price">
              {{ formatPrice(item.price * item.quantity) }}
            </div>
          </div>
        </div>
      </div>

      <div class="order-summary">
        <h2>Order Summary</h2>

        <div class="summary-details">
          <div class="summary-row">
            <span>Subtotal ({{ cartStore.itemCount }} items)</span>
            <span>{{ formatPrice(cartStore.totalPrice) }}</span>
          </div>
          <div class="summary-row">
            <span>Shipping</span>
            <span>Free</span>
          </div>
          <div class="summary-row total">
            <span>Total</span>
            <span>{{ formatPrice(cartStore.totalPrice) }}</span>
          </div>
        </div>

        <div v-if="orderError" class="error-message">
          {{ orderError }}
        </div>

        <div class="actions">
          <button class="btn-back" @click="goBack">Back to Cart</button>
          <button
            class="btn-place-order"
            :disabled="submitting"
            @click="placeOrder"
          >
            {{ submitting ? 'Processing...' : 'Place Order' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.checkout-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

h1 {
  margin-bottom: 2rem;
  color: #333;
}

h2 {
  font-size: 1.2rem;
  margin-bottom: 1rem;
  color: #333;
}

.loading,
.empty {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.checkout-content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 2rem;
}

@media (max-width: 768px) {
  .checkout-content {
    grid-template-columns: 1fr;
  }
}

.order-items {
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 1.5rem;
}

.items-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.order-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid #f0f0f0;
}

.order-item:last-child {
  border-bottom: none;
}

.item-info {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.item-name {
  font-weight: 500;
  color: #333;
}

.item-qty {
  color: #666;
  font-size: 0.9rem;
}

.item-price {
  font-weight: 500;
  color: #333;
}

.order-summary {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 1.5rem;
  height: fit-content;
  position: sticky;
  top: 80px;
}

.summary-details {
  margin-bottom: 1.5rem;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  color: #666;
}

.summary-row.total {
  border-top: 2px solid #ddd;
  margin-top: 0.5rem;
  padding-top: 1rem;
  font-size: 1.2rem;
  font-weight: 600;
  color: #333;
}

.error-message {
  background: #fee;
  color: #c00;
  padding: 0.75rem 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.btn-back {
  width: 100%;
  padding: 0.75rem;
  background: white;
  color: #666;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.95rem;
}

.btn-back:hover {
  background: #f5f5f5;
}

.btn-place-order {
  width: 100%;
  padding: 1rem;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
}

.btn-place-order:hover:not(:disabled) {
  background: #218838;
}

.btn-place-order:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>
