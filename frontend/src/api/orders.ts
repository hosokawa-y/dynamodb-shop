import apiClient from './client'
import type { Order } from './types'

export const ordersApi = {
  async createOrder(): Promise<Order> {
    const response = await apiClient.post<Order>('/orders')
    return response.data
  },

  async getOrders(): Promise<Order[]> {
    const response = await apiClient.get<Order[]>('/orders')
    return response.data
  },

  async getOrderById(orderId: string): Promise<Order> {
    const response = await apiClient.get<Order>(`/orders/${orderId}`)
    return response.data
  },
}
