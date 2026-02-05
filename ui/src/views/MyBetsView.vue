<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useBetsStore } from '@/stores/bets'
import { usePredictionsStore } from '@/stores/predictions'
import { BetStatus } from '@/types/predictions'
import AppHeader from '@/components/AppHeader.vue'
import BottomNav from '@/components/BottomNav.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

const betsStore = useBetsStore()
const predictionsStore = usePredictionsStore()

onMounted(async () => {
  await Promise.all([
    betsStore.fetchBets(),
    predictionsStore.fetchPredictions(),
  ])
})

function getPrediction(predictionId: string) {
  return predictionsStore.predictions.find(p => p.prediction.id === predictionId)
}

function getChoiceName(predictionId: string, choiceId: string) {
  const prediction = getPrediction(predictionId)
  return prediction?.prediction?.choices.find(c => c.id === choiceId)?.name ?? 'Unknown'
}

const betsByStatus = computed(() => {
  const placed = betsStore.sortedBets.filter(b => b.status === BetStatus.Placed)
  const won = betsStore.sortedBets.filter(b => b.status === BetStatus.Won)
  const lost = betsStore.sortedBets.filter(b => b.status === BetStatus.Lost)
  const voided = betsStore.sortedBets.filter(b => b.status === BetStatus.Voided)
  return { placed, won, lost, voided }
})

const stats = computed(() => {
  const total = betsStore.bets.length
  const won = betsByStatus.value.won.length
  const lost = betsByStatus.value.lost.length
  const pending = betsByStatus.value.placed.length
  const totalWon = betsByStatus.value.won.reduce((sum, b) => sum + b.won_amount, 0)
  const totalLost = betsByStatus.value.lost.reduce((sum, b) => sum + b.amount, 0)
  return { total, won, lost, pending, totalWon, totalLost }
})

function statusConfig(status: string) {
  switch (status) {
    case BetStatus.Placed:
      return { label: 'Pending', class: 'bg-primary/20 text-primary' }
    case BetStatus.Won:
      return { label: 'Won', class: 'bg-success/20 text-success' }
    case BetStatus.Lost:
      return { label: 'Lost', class: 'bg-error/20 text-error' }
    case BetStatus.Voided:
      return { label: 'Voided', class: 'bg-gray-500/20 text-gray-400' }
    default:
      return { label: 'Unknown', class: 'bg-gray-500/20 text-gray-400' }
  }
}
</script>

<template>
  <div class="min-h-screen bg-bg pb-20">
    <AppHeader />

    <main class="p-4">
      <h1 class="text-2xl font-bold text-white mb-6">My Bets</h1>

      <div v-if="betsStore.loading || predictionsStore.loading" class="flex justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>

      <div v-else-if="betsStore.error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error">
        {{ betsStore.error }}
      </div>

      <template v-else>
        <!-- Stats -->
        <div class="grid grid-cols-3 gap-3 mb-6">
          <div class="bg-dark-light rounded-xl p-4 text-center">
            <p class="text-2xl font-bold text-primary">{{ stats.pending }}</p>
            <p class="text-xs text-gray-400">Pending</p>
          </div>
          <div class="bg-dark-light rounded-xl p-4 text-center">
            <p class="text-2xl font-bold text-success">{{ stats.won }}</p>
            <p class="text-xs text-gray-400">Won</p>
          </div>
          <div class="bg-dark-light rounded-xl p-4 text-center">
            <p class="text-2xl font-bold text-error">{{ stats.lost }}</p>
            <p class="text-xs text-gray-400">Lost</p>
          </div>
        </div>

        <!-- Profit/Loss -->
        <div v-if="stats.totalWon > 0 || stats.totalLost > 0" class="bg-dark-light rounded-xl p-4 mb-6">
          <div class="flex justify-between items-center">
            <span class="text-gray-400">Net Result</span>
            <span
              class="text-xl font-bold"
              :class="stats.totalWon - stats.totalLost >= 0 ? 'text-success' : 'text-error'"
            >
              {{ stats.totalWon - stats.totalLost >= 0 ? '+' : '' }}{{ stats.totalWon - stats.totalLost }}
            </span>
          </div>
        </div>

        <!-- Bets List -->
        <div class="space-y-3">
          <router-link
            v-for="bet in betsStore.sortedBets"
            :key="bet.id"
            :to="{ name: 'prediction', params: { id: bet.prediction_id } }"
            class="block bg-dark-light rounded-xl p-4 hover:bg-dark-lighter transition-colors"
          >
            <div class="flex items-start justify-between gap-3 mb-2">
              <div class="flex-1 min-w-0">
                <p class="font-medium text-white truncate">
                  {{ getPrediction(bet.prediction_id)?.prediction?.name ?? 'Unknown Prediction' }}
                </p>
                <p class="text-sm text-gray-400">
                  {{ getChoiceName(bet.prediction_id, bet.prediction_choice_id) }}
                </p>
              </div>
              <span
                class="px-2 py-1 rounded-full text-xs font-medium shrink-0"
                :class="statusConfig(bet.status).class"
              >
                {{ statusConfig(bet.status).label }}
              </span>
            </div>

            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-400">Bet: {{ bet.amount }} tokens</span>
              <span v-if="bet.status === BetStatus.Won" class="text-success font-medium">
                +{{ bet.won_amount }} won
              </span>
              <span v-else-if="bet.status === BetStatus.Lost" class="text-error font-medium">
                -{{ bet.amount }} lost
              </span>
            </div>
          </router-link>
        </div>

        <div v-if="betsStore.bets.length === 0" class="text-center py-12">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-dark-light flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" />
            </svg>
          </div>
          <p class="text-gray-400">No bets yet</p>
          <p class="text-gray-500 text-sm mt-1">Place your first bet on a prediction!</p>
          <router-link
            :to="{ name: 'home' }"
            class="inline-block mt-4 bg-primary text-dark px-4 py-2 rounded-lg font-medium"
          >
            View Predictions
          </router-link>
        </div>
      </template>
    </main>

    <BottomNav />
  </div>
</template>
