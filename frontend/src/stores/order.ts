import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ordersApi } from '@/api/orders'
import type { Order } from '@/api/types'

export const useOrderStore = defineStore('order', () => {
  const orders = ref<Order[]>([])
  const currentOrder = ref<Order | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const orderCount = computed(() => orders.value.length)

  async function createOrder(): Promise<Order> {
    loading.value = true
    error.value = null

    try {
      const order = await ordersApi.createOrder()
      // 注文一覧に追加
      orders.value.unshift(order)
      currentOrder.value = order
      return order
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      const errorMessage = err.response?.data?.error || 'Failed to create order'
      error.value = errorMessage
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchOrders() {
    loading.value = true
    error.value = null

    try {
      orders.value = await ordersApi.getOrders()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to fetch orders'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchOrderById(orderId: string) {
    loading.value = true
    error.value = null

    try {
      currentOrder.value = await ordersApi.getOrderById(orderId)
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to fetch order'
      throw e
    } finally {
      loading.value = false
    }
  }

  function clearCurrentOrder() {
    currentOrder.value = null
  }

  return {
    orders,
    currentOrder,
    loading,
    error,
    orderCount,
    createOrder,
    fetchOrders,
    fetchOrderById,
    clearCurrentOrder,
  }
})
