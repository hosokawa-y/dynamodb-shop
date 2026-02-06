<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { useAuthStore } from '@/stores/auth'

const cartStore = useCartStore()
const authStore = useAuthStore()
const router = useRouter()

const updatingItems = ref<Set<string>>(new Set())

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: '/cart' } })
    return
  }
  await cartStore.fetchCart()
})

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}

async function incrementQuantity(productId: string, currentQuantity: number, version: number) {
  if (updatingItems.value.has(productId)) return
  updatingItems.value.add(productId)
  try {
    await cartStore.updateQuantity(productId, currentQuantity + 1, version)
  } catch {
    // エラーはストアで処理済み
  } finally {
    updatingItems.value.delete(productId)
  }
}

async function decrementQuantity(productId: string, currentQuantity: number, version: number) {
  if (updatingItems.value.has(productId)) return
  if (currentQuantity <= 1) {
    await removeItem(productId)
    return
  }
  updatingItems.value.add(productId)
  try {
    await cartStore.updateQuantity(productId, currentQuantity - 1, version)
  } catch {
    // エラーはストアで処理済み
  } finally {
    updatingItems.value.delete(productId)
  }
}

async function removeItem(productId: string) {
  if (updatingItems.value.has(productId)) return
  updatingItems.value.add(productId)
  try {
    await cartStore.removeItem(productId)
  } catch {
    // エラーはストアで処理済み
  } finally {
    updatingItems.value.delete(productId)
  }
}

function goToProducts() {
  router.push('/products')
}
</script>

<template>
  <div class="cart-view">
    <h1>Shopping Cart</h1>

    <div v-if="cartStore.loading && !cartStore.cart" class="loading">Loading cart...</div>

    <div v-else-if="cartStore.error" class="error">
      {{ cartStore.error }}
    </div>

    <div v-else-if="cartStore.itemCount === 0" class="empty-cart">
      <p>Your cart is empty</p>
      <button class="btn-primary" @click="goToProducts">Browse Products</button>
    </div>

    <div v-else class="cart-content">
      <div class="cart-items">
        <div v-for="item in cartStore.items" :key="item.productId" class="cart-item">
          <div class="item-info">
            <h3>{{ item.productName }}</h3>
            <p class="item-price">{{ formatPrice(item.price) }}</p>
          </div>

          <div class="item-quantity">
            <button
              class="qty-btn"
              :disabled="updatingItems.has(item.productId)"
              @click="decrementQuantity(item.productId, item.quantity, item.version)"
            >
              -
            </button>
            <span class="qty-value">{{ item.quantity }}</span>
            <button
              class="qty-btn"
              :disabled="updatingItems.has(item.productId)"
              @click="incrementQuantity(item.productId, item.quantity, item.version)"
            >
              +
            </button>
          </div>

          <div class="item-subtotal">
            {{ formatPrice(item.price * item.quantity) }}
          </div>

          <button
            class="remove-btn"
            :disabled="updatingItems.has(item.productId)"
            @click="removeItem(item.productId)"
          >
            Remove
          </button>
        </div>
      </div>

      <div class="cart-summary">
        <div class="summary-row">
          <span>Items ({{ cartStore.itemCount }})</span>
          <span>{{ formatPrice(cartStore.totalPrice) }}</span>
        </div>
        <div class="summary-row total">
          <span>Total</span>
          <span>{{ formatPrice(cartStore.totalPrice) }}</span>
        </div>
        <button class="btn-checkout" disabled>Proceed to Checkout</button>
        <p class="checkout-note">Checkout will be available in Phase 3</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cart-view {
  max-width: 1000px;
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

.empty-cart {
  text-align: center;
  padding: 3rem;
  background: #f9f9f9;
  border-radius: 8px;
}

.empty-cart p {
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

.cart-content {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 2rem;
}

@media (max-width: 768px) {
  .cart-content {
    grid-template-columns: 1fr;
  }
}

.cart-items {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.cart-item {
  display: grid;
  grid-template-columns: 1fr auto auto auto;
  gap: 1rem;
  align-items: center;
  padding: 1rem;
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
}

@media (max-width: 600px) {
  .cart-item {
    grid-template-columns: 1fr;
    text-align: center;
  }
}

.item-info h3 {
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
  color: #333;
}

.item-price {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.item-quantity {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.qty-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qty-btn:hover:not(:disabled) {
  background: #f5f5f5;
}

.qty-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.qty-value {
  min-width: 2rem;
  text-align: center;
  font-weight: 500;
}

.item-subtotal {
  font-weight: 600;
  color: #333;
  min-width: 80px;
  text-align: right;
}

.remove-btn {
  background: none;
  border: 1px solid #ddd;
  color: #666;
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
}

.remove-btn:hover:not(:disabled) {
  border-color: #c00;
  color: #c00;
}

.remove-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cart-summary {
  background: #f9f9f9;
  padding: 1.5rem;
  border-radius: 8px;
  height: fit-content;
  position: sticky;
  top: 80px;
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

.btn-checkout {
  width: 100%;
  margin-top: 1.5rem;
  padding: 1rem;
  background: #4a90d9;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
}

.btn-checkout:hover:not(:disabled) {
  background: #3a7bc8;
}

.btn-checkout:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.checkout-note {
  margin-top: 0.5rem;
  font-size: 0.8rem;
  color: #999;
  text-align: center;
}
</style>
