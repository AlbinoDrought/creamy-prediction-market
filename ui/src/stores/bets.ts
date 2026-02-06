import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Bet } from '@/types/predictions'
import { api } from '@/api/client'
import { useAuthStore } from './auth'
import { useConfetti } from '@/composables/useConfetti'

export const useBetsStore = defineStore('bets', () => {
  const bets = ref<Bet[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const placingBet = ref(false)
  const newlyWonBets = ref<Bet[]>([])
  const hasInitialized = ref(false)

  const { fire: fireConfetti } = useConfetti()

  const sortedBets = computed(() => {
    return [...bets.value].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  })

  async function swapBets() {
    error.value = null
    try {
      // Track previous bet statuses to detect new wins
      const previousStatuses = new Map(bets.value.map(b => [b.id, b.status]))

      const newBets = await api.getMyBets()

      // Only check for wins after initial load (not on page open)
      if (hasInitialized.value) {
        const newWins = newBets.filter(bet =>
          bet.status === 'won' && previousStatuses.get(bet.id) !== 'won'
        )

        if (newWins.length > 0) {
          newlyWonBets.value = newWins
          // Scale particles based on win amount (10 tokens = 30 particles, 300+ = 200 particles)
          const totalWon = newWins.reduce((sum, bet) => sum + (bet.won_amount ?? 0), 0)
          const minParticles = 30
          const maxParticles = 200
          const minWin = 10
          const maxWin = 300
          const scaled = minParticles + ((totalWon - minWin) / (maxWin - minWin)) * (maxParticles - minParticles)
          const particleCount = Math.round(Math.max(minParticles, Math.min(maxParticles, scaled)))
          fireConfetti(particleCount)
        }
      }

      bets.value = newBets
      hasInitialized.value = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load bets'
    }
  }

  function clearNewlyWonBets() {
    newlyWonBets.value = []
  }

  async function fetchBets() {
    const loadingTimeout = setTimeout(() => {
      loading.value = true
    }, 70)
    try {
      await swapBets()
    } finally {
      clearTimeout(loadingTimeout)
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
    newlyWonBets,
    swapBets,
    fetchBets,
    placeBet,
    increaseBet,
    getBetsForPrediction,
    getBetForPrediction,
    clearNewlyWonBets,
  }
})
