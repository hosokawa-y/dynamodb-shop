<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useProductStore } from '@/stores/product'
import ProductCard from '@/components/ProductCard.vue'

const productStore = useProductStore()
const selectedCategory = ref('')

onMounted(() => {
  productStore.fetchProducts()
})

watch(selectedCategory, (category) => {
  productStore.fetchProducts(category || undefined)
})

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}
</script>

<template>
  <div class="products-page">
    <div class="page-header">
      <h1>Products</h1>

      <div class="filters">
        <select v-model="selectedCategory" class="category-select">
          <option value="">All Categories</option>
          <option v-for="category in productStore.categories" :key="category" :value="category">
            {{ category }}
          </option>
        </select>
      </div>
    </div>

    <div v-if="productStore.loading" class="loading">Loading products...</div>

    <div v-else-if="productStore.error" class="error">
      {{ productStore.error }}
    </div>

    <div v-else-if="productStore.products.length === 0" class="empty">No products found.</div>

    <div v-else class="products-grid">
      <ProductCard
        v-for="product in productStore.products"
        :key="product.id"
        :product="product"
        :format-price="formatPrice"
      />
    </div>
  </div>
</template>

<style scoped>
.products-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-header h1 {
  margin: 0;
  color: #333;
}

.filters {
  display: flex;
  gap: 1rem;
}

.category-select {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  background: white;
  cursor: pointer;
}

.category-select:focus {
  outline: none;
  border-color: #4a90d9;
}

.loading,
.error,
.empty {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.error {
  color: #c00;
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
}
</style>
