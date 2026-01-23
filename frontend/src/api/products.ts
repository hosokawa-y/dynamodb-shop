import apiClient from './client'
import type { CreateProductRequest, Product, UpdateProductRequest } from './types'

export const productsApi = {
  async list(category?: string): Promise<Product[]> {
    const params = category ? { category } : {}
    const response = await apiClient.get<Product[]>('/products', { params })
    return response.data
  },

  async getById(id: string): Promise<Product> {
    const response = await apiClient.get<Product>(`/products/${id}`)
    return response.data
  },

  async create(data: CreateProductRequest): Promise<Product> {
    const response = await apiClient.post<Product>('/products', data)
    return response.data
  },

  async update(id: string, data: UpdateProductRequest): Promise<Product> {
    const response = await apiClient.put<Product>(`/products/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/products/${id}`)
  },
}
