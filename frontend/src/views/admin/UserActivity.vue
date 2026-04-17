<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { activityApi } from '@/api/activity'
import type { UserActivity } from '@/api/types'

const userId = ref('')
const activities = ref<UserActivity[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const actionTypeFilter = ref<string>('')

const actionTypes = [
  { value: '', label: 'All' },
  { value: 'VIEW', label: 'View' },
  { value: 'CLICK', label: 'Click' },
  { value: 'ADD_CART', label: 'Add to Cart' },
  { value: 'PURCHASE', label: 'Purchase' },
]

async function fetchActivities() {
  if (!userId.value.trim()) {
    error.value = 'Please enter a user ID'
    return
  }

  loading.value = true
  error.value = null

  try {
    const params: { limit?: number; actionType?: string } = { limit: 100 }
    if (actionTypeFilter.value) {
      params.actionType = actionTypeFilter.value
    }
    activities.value = await activityApi.getUserActivities(userId.value.trim(), params)
  } catch {
    error.value = 'Failed to load activities'
    activities.value = []
  } finally {
    loading.value = false
  }
}

async function fetchMyActivities() {
  loading.value = true
  error.value = null

  try {
    const params: { limit?: number; actionType?: string } = { limit: 100 }
    if (actionTypeFilter.value) {
      params.actionType = actionTypeFilter.value
    }
    activities.value = await activityApi.getMyActivities(params)
  } catch {
    error.value = 'Failed to load activities'
    activities.value = []
  } finally {
    loading.value = false
  }
}

function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString('ja-JP')
}

function getActionTypeLabel(type: string): string {
  switch (type) {
    case 'VIEW':
      return 'View'
    case 'CLICK':
      return 'Click'
    case 'ADD_CART':
      return 'Add to Cart'
    case 'PURCHASE':
      return 'Purchase'
    default:
      return type
  }
}

function getActionTypeClass(type: string): string {
  switch (type) {
    case 'VIEW':
      return 'type-view'
    case 'CLICK':
      return 'type-click'
    case 'ADD_CART':
      return 'type-add-cart'
    case 'PURCHASE':
      return 'type-purchase'
    default:
      return ''
  }
}

// 統計情報を計算
function getStatistics() {
  const stats = {
    total: activities.value.length,
    view: activities.value.filter((a) => a.actionType === 'VIEW').length,
    click: activities.value.filter((a) => a.actionType === 'CLICK').length,
    addCart: activities.value.filter((a) => a.actionType === 'ADD_CART').length,
    purchase: activities.value.filter((a) => a.actionType === 'PURCHASE').length,
  }
  return stats
}

onMounted(() => {
  // 初期表示時は自分のアクティビティを取得
  fetchMyActivities()
})
</script>

<template>
  <div class="user-activity">
    <h1>User Activity Analytics</h1>

    <!-- Search Form -->
    <div class="search-section">
      <div class="search-form">
        <div class="form-group">
          <label>User ID:</label>
          <input v-model="userId" type="text" placeholder="Enter user ID..." @keyup.enter="fetchActivities" />
        </div>
        <div class="form-group">
          <label>Action Type:</label>
          <select v-model="actionTypeFilter">
            <option v-for="type in actionTypes" :key="type.value" :value="type.value">
              {{ type.label }}
            </option>
          </select>
        </div>
        <div class="button-group">
          <button @click="fetchActivities" :disabled="loading">Search User</button>
          <button @click="fetchMyActivities" :disabled="loading" class="btn-secondary">My Activities</button>
        </div>
      </div>
    </div>

    <!-- Statistics -->
    <div v-if="activities.length > 0" class="statistics">
      <h2>Statistics</h2>
      <div class="stat-cards">
        <div class="stat-card">
          <span class="stat-value">{{ getStatistics().total }}</span>
          <span class="stat-label">Total</span>
        </div>
        <div class="stat-card type-view">
          <span class="stat-value">{{ getStatistics().view }}</span>
          <span class="stat-label">Views</span>
        </div>
        <div class="stat-card type-click">
          <span class="stat-value">{{ getStatistics().click }}</span>
          <span class="stat-label">Clicks</span>
        </div>
        <div class="stat-card type-add-cart">
          <span class="stat-value">{{ getStatistics().addCart }}</span>
          <span class="stat-label">Add to Cart</span>
        </div>
        <div class="stat-card type-purchase">
          <span class="stat-value">{{ getStatistics().purchase }}</span>
          <span class="stat-label">Purchases</span>
        </div>
      </div>
    </div>

    <!-- Activity List -->
    <div class="activity-section">
      <h2>Activity Log</h2>

      <div v-if="loading" class="loading">Loading activities...</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <div v-else-if="activities.length === 0" class="no-data">No activities found</div>

      <table v-else class="activity-table">
        <thead>
          <tr>
            <th>Date</th>
            <th>User ID</th>
            <th>Action</th>
            <th>Product ID</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(activity, index) in activities" :key="index">
            <td>{{ formatDateTime(activity.timestamp) }}</td>
            <td class="user-id">{{ activity.userId }}</td>
            <td>
              <span :class="['type-badge', getActionTypeClass(activity.actionType)]">
                {{ getActionTypeLabel(activity.actionType) }}
              </span>
            </td>
            <td class="product-id">{{ activity.productId || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.user-activity {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.user-activity h1 {
  margin: 0 0 1.5rem;
  color: #333;
}

.user-activity h2 {
  margin: 0 0 1rem;
  font-size: 1.1rem;
  color: #333;
}

.loading,
.error,
.no-data {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error {
  color: #c00;
}

/* Search Section */
.search-section {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.search-form {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
  flex-wrap: wrap;
}

.form-group {
  flex: 1;
  min-width: 200px;
}

.form-group label {
  display: block;
  font-size: 0.9rem;
  color: #666;
  margin-bottom: 0.25rem;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.button-group {
  display: flex;
  gap: 0.5rem;
}

button {
  padding: 0.5rem 1rem;
  background: #4a90d9;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 0.2s;
}

button:hover:not(:disabled) {
  background: #3a7bc8;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

button.btn-secondary {
  background: white;
  color: #4a90d9;
  border: 1px solid #4a90d9;
}

button.btn-secondary:hover:not(:disabled) {
  background: #f0f7ff;
}

/* Statistics */
.statistics {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.stat-cards {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.stat-card {
  flex: 1;
  min-width: 100px;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 1rem;
  text-align: center;
}

.stat-card.type-view {
  background: #e3f2fd;
}

.stat-card.type-click {
  background: #fff3e0;
}

.stat-card.type-add-cart {
  background: #e8f5e9;
}

.stat-card.type-purchase {
  background: #fce4ec;
}

.stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: bold;
  color: #333;
}

.stat-label {
  display: block;
  font-size: 0.85rem;
  color: #666;
  margin-top: 0.25rem;
}

/* Activity Section */
.activity-section {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
}

.activity-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.activity-table th,
.activity-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.activity-table th {
  color: #666;
  font-weight: 500;
  background: #f9f9f9;
}

.user-id,
.product-id {
  font-family: monospace;
  font-size: 0.85rem;
  color: #666;
}

.type-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.type-view {
  background: #e3f2fd;
  color: #1565c0;
}

.type-click {
  background: #fff3e0;
  color: #e65100;
}

.type-add-cart {
  background: #e8f5e9;
  color: #2e7d32;
}

.type-purchase {
  background: #fce4ec;
  color: #c2185b;
}

@media (max-width: 768px) {
  .search-form {
    flex-direction: column;
  }

  .form-group {
    width: 100%;
  }

  .stat-cards {
    flex-direction: column;
  }

  .activity-table {
    font-size: 0.8rem;
  }

  .activity-table th,
  .activity-table td {
    padding: 0.5rem;
  }
}
</style>
