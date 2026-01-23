import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { productsApi } from '@/api'
import type { Product, CreateProductRequest, UpdateProductRequest } from '@/api/types'

export const useProductStore = defineStore('product', () => {
  const products = ref<Product[]>([])
  const currentProduct = ref<Product | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const categories = computed(() => {
    const cats = new Set(products.value.map((p) => p.category))
    return Array.from(cats).sort()
  })

  async function fetchProducts(category?: string) {
    loading.value = true
    error.value = null

    try {
      products.value = await productsApi.list(category)
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to fetch products'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchProductById(id: string) {
    loading.value = true
    error.value = null

    try {
      currentProduct.value = await productsApi.getById(id)
      return currentProduct.value
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to fetch product'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function createProduct(data: CreateProductRequest) {
    loading.value = true
    error.value = null

    try {
      const product = await productsApi.create(data)
      products.value.push(product)
      return product
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to create product'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateProduct(id: string, data: UpdateProductRequest) {
    loading.value = true
    error.value = null

    try {
      const product = await productsApi.update(id, data)
      const index = products.value.findIndex((p) => p.id === id)
      if (index !== -1) {
        products.value[index] = product
      }
      if (currentProduct.value?.id === id) {
        currentProduct.value = product
      }
      return product
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to update product'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteProduct(id: string) {
    loading.value = true
    error.value = null

    try {
      await productsApi.delete(id)
      products.value = products.value.filter((p) => p.id !== id)
      if (currentProduct.value?.id === id) {
        currentProduct.value = null
      }
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to delete product'
      throw e
    } finally {
      loading.value = false
    }
  }

  function clearCurrentProduct() {
    currentProduct.value = null
  }

  return {
    products,
    currentProduct,
    loading,
    error,
    categories,
    fetchProducts,
    fetchProductById,
    createProduct,
    updateProduct,
    deleteProduct,
    clearCurrentProduct,
  }
})
