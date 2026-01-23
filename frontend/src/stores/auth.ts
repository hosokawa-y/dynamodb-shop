import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import type { User, LoginRequest, RegisterRequest } from '@/api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  // 初期化: localStorageから復元
  function initialize() {
    const storedToken = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')

    if (storedToken) {
      token.value = storedToken
    }
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        localStorage.removeItem('user')
      }
    }
  }

  async function register(data: RegisterRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await authApi.register(data)
      token.value = response.token
      user.value = response.user

      localStorage.setItem('token', response.token)
      localStorage.setItem('user', JSON.stringify(response.user))

      return response
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Registration failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function login(data: LoginRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await authApi.login(data)
      token.value = response.token
      user.value = response.user

      localStorage.setItem('token', response.token)
      localStorage.setItem('user', JSON.stringify(response.user))

      return response
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Login failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    if (!token.value) return

    loading.value = true
    error.value = null

    try {
      const userData = await authApi.getProfile()
      user.value = userData
      localStorage.setItem('user', JSON.stringify(userData))
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      error.value = err.response?.data?.error || 'Failed to fetch profile'
      throw e
    } finally {
      loading.value = false
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    initialize,
    register,
    login,
    fetchProfile,
    logout,
  }
})
