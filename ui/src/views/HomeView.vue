<script setup lang="ts">
import { onMounted } from 'vue'
import { usePredictionsStore } from '@/stores/predictions'
import { PredictionStatus } from '@/types/predictions'
import AppHeader from '@/components/AppHeader.vue'
import BottomNav from '@/components/BottomNav.vue'
import PredictionCard from '@/components/PredictionCard.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

const predictionsStore = usePredictionsStore()

onMounted(() => {
  predictionsStore.fetchPredictions()
})

function filterByStatus(status: string) {
  if (status === 'all') return predictionsStore.predictions
  return predictionsStore.predictions.filter(p => p.prediction.status === status)
}

const sections = [
  { title: 'Open Predictions', status: PredictionStatus.Open, empty: 'No open predictions' },
  { title: 'Closed (Awaiting Result)', status: PredictionStatus.Closed, empty: 'No closed predictions' },
  { title: 'Decided', status: PredictionStatus.Decided, empty: 'No decided predictions' },
]
</script>

<template>
  <div class="min-h-screen bg-bg pb-20">
    <AppHeader />

    <main class="p-4 space-y-6">
      <div v-if="predictionsStore.loading" class="flex justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>

      <div v-else-if="predictionsStore.error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error">
        {{ predictionsStore.error }}
      </div>

      <template v-else>
        <div v-for="section in sections" :key="section.status">
          <template v-if="filterByStatus(section.status).length > 0">
            <h2 class="text-lg font-semibold text-white mb-3">{{ section.title }}</h2>
            <div class="space-y-3">
              <PredictionCard
                v-for="prediction in filterByStatus(section.status)"
                :key="prediction.prediction.id"
                :prediction="prediction.prediction"
              />
            </div>
          </template>
        </div>

        <div v-if="predictionsStore.predictions.length === 0" class="text-center py-12">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-dark-light flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          </div>
          <p class="text-gray-400">No predictions yet</p>
          <p class="text-gray-500 text-sm mt-1">Check back soon for Superbowl bets!</p>
        </div>
      </template>
    </main>

    <BottomNav />
  </div>
</template>
