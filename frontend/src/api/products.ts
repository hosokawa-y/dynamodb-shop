import apiClient from './client'
import type {
  CreateProductRequest,
  Product,
  UpdateProductRequest,
  PriceHistory,
  InventoryLog,
  UpdatePriceRequest,
  AdjustStockRequest,
} from './types'

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

  // 価格履歴API
  async getPriceHistory(
    id: string,
    params?: { limit?: number; start?: string; end?: string },
  ): Promise<PriceHistory[]> {
    const response = await apiClient.get<PriceHistory[]>(`/products/${id}/price-history`, { params })
    return response.data
  },

  async updatePrice(id: string, data: UpdatePriceRequest): Promise<void> {
    await apiClient.put(`/products/${id}/price`, data)
  },

  // 在庫管理API
  async getInventoryLogs(
    id: string,
    params?: { limit?: number; start?: string; end?: string },
  ): Promise<InventoryLog[]> {
    const response = await apiClient.get<InventoryLog[]>(`/products/${id}/inventory-logs`, { params })
    return response.data
  },

  async adjustStock(id: string, data: AdjustStockRequest): Promise<void> {
    await apiClient.put(`/products/${id}/stock`, data)
  },

  // 管理者用：在庫変動履歴（全商品）
  async getAdminInventoryLogs(params: { productId: string; limit?: number }): Promise<InventoryLog[]> {
    const response = await apiClient.get<InventoryLog[]>('/admin/inventory-logs', { params })
    return response.data
  },
}
