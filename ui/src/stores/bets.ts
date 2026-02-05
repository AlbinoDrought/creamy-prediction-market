import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Bet } from '@/types/predictions'
import { api } from '@/api/client'
import { useAuthStore } from './auth'

export const useBetsStore = defineStore('bets', () => {
  const bets = ref<Bet[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const placingBet = ref(false)

  const sortedBets = computed(() => {
    return [...bets.value].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  })

  async function fetchBets() {
    loading.value = true
    error.value = null
    try {
      bets.value = await api.getMyBets()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load bets'
    } finally {
      loading.value = false
    }
  }

  async function placeBet(predictionId: string, choiceId: string, amount: number) {
    placingBet.value = true
    error.value = null
    try {
      const bet = await api.placeBet(predictionId, choiceId, amount)
      bets.value.push(bet)

      // Update user's token balance
      const authStore = useAuthStore()
      if (authStore.user) {
        authStore.updateTokens(authStore.user.tokens - amount)
      }

      return bet
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to place bet'
      throw e
    } finally {
      placingBet.value = false
    }
  }

  function getBetsForPrediction(predictionId: string) {
    return bets.value.filter(bet => bet.prediction_id === predictionId)
  }

  return {
    bets,
    sortedBets,
    loading,
    error,
    placingBet,
    fetchBets,
    placeBet,
    getBetsForPrediction,
  }
})
