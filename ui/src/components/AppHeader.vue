<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

function goToAdmin() {
  router.push({ name: 'admin-dashboard' })
}
</script>

<template>
  <header class="bg-dark px-4 py-3 flex items-center justify-between border-b border-dark-lighter">
    <div class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-secondary flex items-center justify-center font-bold text-dark">
        {{ authStore.user?.name.charAt(0).toUpperCase() }}
      </div>
      <div>
        <p class="font-medium">{{ authStore.user?.name }}</p>
        <p class="text-sm text-gray-400">Player</p>
      </div>
    </div>
    <div class="flex items-center gap-3">
      <div class="bg-dark-light px-4 py-2 rounded-full flex items-center gap-2">
        <span class="text-primary font-bold">{{ authStore.user?.tokens ?? 0 }}</span>
        <span class="text-gray-400 text-sm">tokens</span>
      </div>
      <button
        v-if="authStore.isAdmin"
        @click="goToAdmin"
        class="bg-secondary hover:bg-secondary-dark px-3 py-2 rounded-lg text-sm font-medium transition-colors"
      >
        Admin
      </button>
    </div>
  </header>
</template>
