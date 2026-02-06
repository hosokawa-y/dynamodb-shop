import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { cartApi } from '@/api'
import type { Cart, CartItem, AddToCartRequest, UpdateCartRequest } from '@/api/types'

export const useCartStore = defineStore('cart', () => {
  const cart = ref<Cart | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const itemCount = computed(() => cart.value?.itemCount ?? 0)
  const totalPrice = computed(() => cart.value?.totalPrice ?? 0)
  const items = computed(() => cart.value?.items ?? [])

  async function fetchCart() {
    loading.value = true
    error.value = null

    try {
      cart.value = await cartApi.getCart()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to fetch cart'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function addItem(data: AddToCartRequest) {
    loading.value = true
    error.value = null

    try {
      await cartApi.addItem(data)
      // カート全体を再取得して最新状態に更新
      await fetchCart()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to add item to cart'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateQuantity(productId: string, quantity: number, version: number) {
    loading.value = true
    error.value = null

    try {
      const data: UpdateCartRequest = { quantity, version }
      await cartApi.updateQuantity(productId, data)
      // カート全体を再取得して最新状態に更新
      await fetchCart()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      // 楽観的ロックによる競合エラー
      if (err.response?.data?.error?.includes('concurrent')) {
        error.value = 'Item was modified. Please try again.'
        // 最新データを取得
        await fetchCart()
      } else {
        error.value = err.response?.data?.error || 'Failed to update quantity'
      }
      throw e
    } finally {
      loading.value = false
    }
  }

  async function removeItem(productId: string) {
    loading.value = true
    error.value = null

    try {
      await cartApi.removeItem(productId)
      // カート全体を再取得して最新状態に更新
      await fetchCart()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to remove item'
      throw e
    } finally {
      loading.value = false
    }
  }

  function clearCart() {
    cart.value = null
  }

  function getItemByProductId(productId: string): CartItem | undefined {
    return cart.value?.items.find((item) => item.productId === productId)
  }

  return {
    cart,
    loading,
    error,
    itemCount,
    totalPrice,
    items,
    fetchCart,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
    getItemByProductId,
  }
})
