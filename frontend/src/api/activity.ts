import apiClient from './client'
import type { UserActivity, LogActivityRequest } from './types'

export const activityApi = {
  // 行動ログを1件記録
  async log(data: LogActivityRequest): Promise<void> {
    await apiClient.post('/activity', data)
  },

  // 複数の行動ログを一括記録
  async logBatch(data: LogActivityRequest[]): Promise<void> {
    await apiClient.post('/activity/batch', data)
  },

  // 自分の行動ログを取得
  async getMyActivities(params?: { limit?: number; actionType?: string }): Promise<UserActivity[]> {
    const response = await apiClient.get<UserActivity[]>('/activity', { params })
    return response.data
  },

  // 管理者用：特定ユーザーの行動ログを取得
  async getUserActivities(
    userId: string,
    params?: { limit?: number; actionType?: string },
  ): Promise<UserActivity[]> {
    const response = await apiClient.get<UserActivity[]>(`/admin/users/${userId}/activities`, {
      params,
    })
    return response.data
  },
}
