<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const cartStore = useCartStore()
const router = useRouter()

// ログイン時にカートを取得
watch(
  () => authStore.isAuthenticated,
  async (isAuthenticated) => {
    if (isAuthenticated) {
      try {
        await cartStore.fetchCart()
      } catch {
        // エラーは無視（ログイン直後など）
      }
    } else {
      cartStore.clearCart()
    }
  },
)

onMounted(async () => {
  if (authStore.isAuthenticated) {
    try {
      await cartStore.fetchCart()
    } catch {
      // エラーは無視
    }
  }
})

function handleLogout() {
  cartStore.clearCart()
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <header class="app-header">
    <div class="header-content">
      <RouterLink to="/" class="logo">DynamoDB Shop</RouterLink>

      <nav class="nav-links">
        <RouterLink to="/products">Products</RouterLink>

        <template v-if="authStore.isAuthenticated">
          <RouterLink to="/cart" class="cart-link">
            Cart
            <span v-if="cartStore.itemCount > 0" class="cart-badge">{{ cartStore.itemCount }}</span>
          </RouterLink>
          <RouterLink to="/profile">{{ authStore.user?.name || 'Profile' }}</RouterLink>
          <button class="logout-btn" @click="handleLogout">Logout</button>
        </template>

        <template v-else>
          <RouterLink to="/login">Login</RouterLink>
          <RouterLink to="/register" class="register-link">Register</RouterLink>
        </template>
      </nav>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo {
  font-size: 1.25rem;
  font-weight: bold;
  color: #4a90d9;
  text-decoration: none;
}

.logo:hover {
  color: #3a7bc8;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.nav-links a {
  color: #555;
  text-decoration: none;
  transition: color 0.2s;
}

.nav-links a:hover {
  color: #4a90d9;
}

.nav-links a.router-link-active {
  color: #4a90d9;
  font-weight: 500;
}

.register-link {
  background: #4a90d9;
  color: white !important;
  padding: 0.5rem 1rem;
  border-radius: 4px;
}

.register-link:hover {
  background: #3a7bc8;
}

.logout-btn {
  background: none;
  border: 1px solid #ddd;
  color: #666;
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition:
    border-color 0.2s,
    color 0.2s;
}

.logout-btn:hover {
  border-color: #c00;
  color: #c00;
}

.cart-link {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.cart-badge {
  background: #e53935;
  color: white;
  font-size: 0.7rem;
  font-weight: bold;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}
</style>
