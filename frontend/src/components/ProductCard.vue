<script setup lang="ts">
import type { Product } from '@/api/types'

defineProps<{
  product: Product
  formatPrice: (price: number) => string
}>()
</script>

<template>
  <RouterLink :to="`/products/${product.id}`" class="product-card">
    <div class="product-image">
      <img v-if="product.imageUrl" :src="product.imageUrl" :alt="product.name" />
      <div v-else class="no-image">No Image</div>
    </div>
    <div class="product-content">
      <span class="category">{{ product.category }}</span>
      <h3 class="name">{{ product.name }}</h3>
      <p class="price">{{ formatPrice(product.price) }}</p>
      <p class="stock" :class="{ 'out-of-stock': product.stock === 0 }">
        {{ product.stock > 0 ? `Stock: ${product.stock}` : 'Out of Stock' }}
      </p>
    </div>
  </RouterLink>
</template>

<style scoped>
.product-card {
  display: block;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  text-decoration: none;
  color: inherit;
  transition:
    transform 0.2s,
    box-shadow 0.2s;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.product-image {
  aspect-ratio: 1;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  color: #999;
}

.product-content {
  padding: 1rem;
}

.category {
  display: inline-block;
  background: #e8f4fd;
  color: #4a90d9;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
}

.name {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
  color: #333;
  line-height: 1.3;
}

.price {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: bold;
  color: #e74c3c;
}

.stock {
  margin: 0;
  font-size: 0.85rem;
  color: #27ae60;
}

.stock.out-of-stock {
  color: #c00;
}
</style>
