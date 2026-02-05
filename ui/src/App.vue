<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useSSE } from '@/composables/useSSE'

const authStore = useAuthStore()
const { connect } = useSSE()

onMounted(async () => {
  // Try to restore session on app load
  if (authStore.token && !authStore.user) {
    await authStore.fetchUser()
  }
})

// Reconnect SSE when auth state changes to pick up user ID
watch(() => authStore.token, () => {
  connect()
})
</script>

<template>
  <RouterView />
</template>
