<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

function handleLogout() {
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <div class="profile-page">
    <div class="profile-card">
      <h1>Profile</h1>

      <div v-if="authStore.user" class="profile-info">
        <div class="info-row">
          <span class="label">Name</span>
          <span class="value">{{ authStore.user.name }}</span>
        </div>

        <div class="info-row">
          <span class="label">Email</span>
          <span class="value">{{ authStore.user.email }}</span>
        </div>

        <div class="info-row">
          <span class="label">Member since</span>
          <span class="value">{{ formatDate(authStore.user.createdAt) }}</span>
        </div>
      </div>

      <div class="actions">
        <button class="btn-danger" @click="handleLogout">Logout</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  max-width: 600px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.profile-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 2rem;
}

.profile-card h1 {
  margin: 0 0 1.5rem;
  color: #333;
  border-bottom: 1px solid #eee;
  padding-bottom: 1rem;
}

.profile-info {
  margin-bottom: 2rem;
}

.info-row {
  display: flex;
  padding: 0.75rem 0;
  border-bottom: 1px solid #f5f5f5;
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  width: 120px;
  color: #666;
  font-weight: 500;
}

.value {
  flex: 1;
  color: #333;
}

.actions {
  padding-top: 1rem;
  border-top: 1px solid #eee;
}

.btn-danger {
  padding: 0.75rem 1.5rem;
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-danger:hover {
  background: #c0392b;
}
</style>
