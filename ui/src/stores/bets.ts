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

  async function swapBets() {
    error.value = null
    try {
      bets.value = await api.getMyBets()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load bets'
    }
  }

  async function fetchBets() {
    loading.value = true
    try {
      await swapBets()
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

  async function increaseBet(betId: string, newAmount: number) {
    placingBet.value = true
    error.value = null
    try {
      const existingBet = bets.value.find(b => b.id === betId)
      if (!existingBet) throw new Error('Bet not found')

      const additionalTokens = newAmount - existingBet.amount
      if (additionalTokens <= 0) throw new Error('New amount must be greater than current amount')

      const updatedBet = await api.updateBetAmount(betId, newAmount)

      // Update the bet in our local state
      const index = bets.value.findIndex(b => b.id === betId)
      if (index !== -1) {
        bets.value[index] = updatedBet
      }

      // Update user's token balance (only deduct the additional amount)
      const authStore = useAuthStore()
      if (authStore.user) {
        authStore.updateTokens(authStore.user.tokens - additionalTokens)
      }

      return updatedBet
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to increase bet'
      throw e
    } finally {
      placingBet.value = false
    }
  }

  function getBetsForPrediction(predictionId: string) {
    return bets.value.filter(bet => bet.prediction_id === predictionId)
  }

  function getBetForPrediction(predictionId: string) {
    return bets.value.find(bet => bet.prediction_id === predictionId)
  }

  return {
    bets,
    sortedBets,
    loading,
    error,
    placingBet,
    swapBets,
    fetchBets,
    placeBet,
    increaseBet,
    getBetsForPrediction,
    getBetForPrediction,
  }
})
