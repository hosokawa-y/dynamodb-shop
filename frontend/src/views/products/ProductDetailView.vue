<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  id: string
}>()

const router = useRouter()
const productStore = useProductStore()
const authStore = useAuthStore()

onMounted(() => {
  productStore.fetchProductById(props.id)
})

onUnmounted(() => {
  productStore.clearCurrentProduct()
})

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('ja-JP')
}

function goBack() {
  router.push('/products')
}
</script>

<template>
  <div class="product-detail-page">
    <button class="back-button" @click="goBack">&larr; Back to Products</button>

    <div v-if="productStore.loading" class="loading">Loading product...</div>

    <div v-else-if="productStore.error" class="error">
      {{ productStore.error }}
    </div>

    <div v-else-if="productStore.currentProduct" class="product-detail">
      <div class="product-image">
        <img
          v-if="productStore.currentProduct.imageUrl"
          :src="productStore.currentProduct.imageUrl"
          :alt="productStore.currentProduct.name"
        />
        <div v-else class="no-image">No Image</div>
      </div>

      <div class="product-info">
        <span class="category-badge">{{ productStore.currentProduct.category }}</span>
        <h1>{{ productStore.currentProduct.name }}</h1>
        <p class="price">{{ formatPrice(productStore.currentProduct.price) }}</p>
        <p class="description">{{ productStore.currentProduct.description }}</p>

        <div class="stock-info">
          <span :class="['stock', productStore.currentProduct.stock > 0 ? 'in-stock' : 'out-of-stock']">
            {{ productStore.currentProduct.stock > 0 ? `In Stock (${productStore.currentProduct.stock})` : 'Out of Stock' }}
          </span>
        </div>

        <div v-if="authStore.isAuthenticated && productStore.currentProduct.stock > 0" class="actions">
          <button class="btn-primary">Add to Cart</button>
        </div>

        <div v-else-if="!authStore.isAuthenticated" class="login-prompt">
          <RouterLink to="/login">Login</RouterLink> to add items to your cart
        </div>

        <div class="meta">
          <p>Created: {{ formatDate(productStore.currentProduct.createdAt) }}</p>
          <p>Updated: {{ formatDate(productStore.currentProduct.updatedAt) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.product-detail-page {
  max-width: 1000px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.back-button {
  background: none;
  border: none;
  color: #4a90d9;
  font-size: 1rem;
  cursor: pointer;
  padding: 0;
  margin-bottom: 1.5rem;
}

.back-button:hover {
  text-decoration: underline;
}

.loading,
.error {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.error {
  color: #c00;
}

.product-detail {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

@media (max-width: 768px) {
  .product-detail {
    grid-template-columns: 1fr;
  }
}

.product-image {
  aspect-ratio: 1;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  color: #999;
  font-size: 1.2rem;
}

.product-info {
  padding: 2rem;
}

.category-badge {
  display: inline-block;
  background: #e8f4fd;
  color: #4a90d9;
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
}

.product-info h1 {
  margin: 0 0 0.5rem;
  color: #333;
}

.price {
  font-size: 1.75rem;
  font-weight: bold;
  color: #e74c3c;
  margin: 0 0 1rem;
}

.description {
  color: #666;
  line-height: 1.6;
  margin-bottom: 1.5rem;
}

.stock-info {
  margin-bottom: 1.5rem;
}

.stock {
  font-weight: 500;
}

.in-stock {
  color: #27ae60;
}

.out-of-stock {
  color: #c00;
}

.actions {
  margin-bottom: 1.5rem;
}

.btn-primary {
  padding: 0.75rem 2rem;
  background: #4a90d9;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #3a7bc8;
}

.login-prompt {
  background: #f5f5f5;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1.5rem;
  color: #666;
}

.login-prompt a {
  color: #4a90d9;
  text-decoration: none;
}

.login-prompt a:hover {
  text-decoration: underline;
}

.meta {
  border-top: 1px solid #eee;
  padding-top: 1rem;
  color: #999;
  font-size: 0.85rem;
}

.meta p {
  margin: 0.25rem 0;
}
</style>
