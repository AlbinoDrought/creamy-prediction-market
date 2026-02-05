<script setup lang="ts">
import { onMounted, watch, ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useBetsStore } from '@/stores/bets'
import { useSSE } from '@/composables/useSSE'
import Toast from '@/components/Toast.vue'

const authStore = useAuthStore()
const betsStore = useBetsStore()
const { connect } = useSSE()

const showWinToast = ref(false)
const winMessage = ref('')

const totalWinnings = computed(() => {
  return betsStore.newlyWonBets.reduce((sum, bet) => sum + (bet.won_amount ?? 0), 0)
})

watch(() => betsStore.newlyWonBets, (newWins) => {
  if (newWins.length > 0) {
    const totalWon = newWins.reduce((sum, bet) => sum + (bet.won_amount ?? 0), 0)
    winMessage.value = `You won ${totalWon} tokens!`
    showWinToast.value = true
  }
}, { deep: true })

function closeWinToast() {
  showWinToast.value = false
  betsStore.clearNewlyWonBets()
}

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
  <Toast
    v-if="showWinToast"
    :message="winMessage"
    type="success"
    :duration="5000"
    @close="closeWinToast"
  />
</template>
