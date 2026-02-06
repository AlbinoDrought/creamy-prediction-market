<script setup lang="ts">
import { onMounted, watch, ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useBetsStore } from '@/stores/bets'
import { useAchievementsStore } from '@/stores/achievements'
import { useShopStore } from '@/stores/shop'
import { useSSE } from '@/composables/useSSE'
import Toast from '@/components/Toast.vue'
import AchievementUnlock from '@/components/AchievementUnlock.vue'

const authStore = useAuthStore()
const betsStore = useBetsStore()
const achievementsStore = useAchievementsStore()
const shopStore = useShopStore()
const { connect } = useSSE()

const showWinToast = ref(false)
const winMessage = ref('')
const showAchievementToast = ref(false)

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

watch(() => achievementsStore.newlyEarnedAchievement, (achievement) => {
  if (achievement) {
    showAchievementToast.value = true
  }
})

function closeWinToast() {
  showWinToast.value = false
  betsStore.clearNewlyWonBets()
}

function closeAchievementToast() {
  showAchievementToast.value = false
  achievementsStore.clearNewlyEarned()
}

onMounted(async () => {
  // Fetch all achievements (public, needed for displaying icons)
  achievementsStore.fetchAchievements()

  // Fetch shop items (public, needed for shop display)
  shopStore.fetchShopItems()

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
  <AchievementUnlock
    v-if="showAchievementToast && achievementsStore.newlyEarnedAchievement"
    :achievement="achievementsStore.newlyEarnedAchievement"
    :duration="5000"
    @close="closeAchievementToast"
  />
</template>
