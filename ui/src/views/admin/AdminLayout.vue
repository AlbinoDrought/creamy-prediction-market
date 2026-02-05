<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'
import { computed } from 'vue'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const navItems = [
  { name: 'admin-dashboard', label: 'Predictions', icon: 'list' },
  { name: 'admin-users', label: 'Users', icon: 'users' },
]

const currentRoute = computed(() => route.name)

function goToHome() {
  router.push({ name: 'home' })
}

function logout() {
  authStore.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen bg-bg">
    <!-- Admin Header -->
    <header class="bg-secondary px-4 py-3 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <h1 class="text-lg font-bold text-white">Admin Panel</h1>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-white/80 text-sm">{{ authStore.user?.name }}</span>
        <button
          @click="logout"
          class="text-white/80 hover:text-white text-sm"
        >
          Logout
        </button>
      </div>
    </header>

    <!-- Admin Nav -->
    <nav class="bg-dark-light border-b border-dark-lighter">
      <div class="flex">
        <router-link
          v-for="item in navItems"
          :key="item.name"
          :to="{ name: item.name }"
          class="flex-1 py-3 px-4 text-center text-sm font-medium transition-colors border-b-2"
          :class="currentRoute === item.name
            ? 'text-secondary border-secondary'
            : 'text-gray-400 border-transparent hover:text-white'"
        >
          {{ item.label }}
        </router-link>
      </div>
    </nav>

    <!-- Content -->
    <main class="p-4">
      <router-view />
    </main>
  </div>
</template>
