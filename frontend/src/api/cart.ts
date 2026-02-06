import apiClient from './client'
import type { Cart, CartItem, AddToCartRequest, UpdateCartRequest } from './types'

export const cartApi = {
  async getCart(): Promise<Cart> {
    const response = await apiClient.get<Cart>('/cart')
    return response.data
  },

  async addItem(data: AddToCartRequest): Promise<CartItem> {
    const response = await apiClient.post<CartItem>('/cart/items', data)
    return response.data
  },

  async updateQuantity(productId: string, data: UpdateCartRequest): Promise<CartItem> {
    const response = await apiClient.put<CartItem>(`/cart/items/${productId}`, data)
    return response.data
  },

  async removeItem(productId: string): Promise<void> {
    await apiClient.delete(`/cart/items/${productId}`)
  },
}
