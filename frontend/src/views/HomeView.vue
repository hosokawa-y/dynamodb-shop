<script setup lang="ts">
import { onMounted } from 'vue'
import { useProductStore } from '@/stores/product'
import { useAuthStore } from '@/stores/auth'
import ProductCard from '@/components/ProductCard.vue'

const productStore = useProductStore()
const authStore = useAuthStore()

onMounted(() => {
  productStore.fetchProducts()
})

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}
</script>

<template>
  <div class="home-page">
    <section class="hero">
      <h1>Welcome to DynamoDB Shop</h1>
      <p>Discover our amazing products</p>
      <div class="hero-actions">
        <RouterLink to="/products" class="btn-primary">Browse Products</RouterLink>
        <RouterLink v-if="!authStore.isAuthenticated" to="/register" class="btn-secondary">
          Create Account
        </RouterLink>
      </div>
    </section>

    <section class="featured-products">
      <h2>Featured Products</h2>

      <div v-if="productStore.loading" class="loading">Loading products...</div>

      <div v-else-if="productStore.products.length > 0" class="products-grid">
        <ProductCard
          v-for="product in productStore.products.slice(0, 4)"
          :key="product.id"
          :product="product"
          :format-price="formatPrice"
        />
      </div>

      <div v-else class="empty">No products available.</div>

      <div v-if="productStore.products.length > 4" class="view-all">
        <RouterLink to="/products">View All Products &rarr;</RouterLink>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.hero {
  text-align: center;
  padding: 4rem 1rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  color: white;
  margin: 2rem 0;
}

.hero h1 {
  margin: 0 0 0.5rem;
  font-size: 2.5rem;
}

.hero p {
  margin: 0 0 2rem;
  font-size: 1.2rem;
  opacity: 0.9;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn-primary,
.btn-secondary {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  text-decoration: none;
  font-weight: 500;
  transition:
    background 0.2s,
    transform 0.2s;
}

.btn-primary {
  background: white;
  color: #667eea;
}

.btn-primary:hover {
  transform: translateY(-2px);
}

.btn-secondary {
  background: transparent;
  color: white;
  border: 2px solid white;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.1);
}

.featured-products {
  padding: 2rem 0;
}

.featured-products h2 {
  margin: 0 0 1.5rem;
  color: #333;
}

.loading,
.empty {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
}

.view-all {
  text-align: center;
  margin-top: 2rem;
}

.view-all a {
  color: #4a90d9;
  text-decoration: none;
  font-weight: 500;
}

.view-all a:hover {
  text-decoration: underline;
}
</style>
