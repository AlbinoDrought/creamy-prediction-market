<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

onMounted(async () => {
  // Try to restore session on app load
  if (authStore.token && !authStore.user) {
    await authStore.fetchUser()
  }
})
</script>

<template>
  <RouterView />
</template>
