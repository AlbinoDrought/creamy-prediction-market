import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PredictionWithOdds } from '@/types/predictions'
import { api } from '@/api/client'

export const usePredictionsStore = defineStore('predictions', () => {
  const predictions = ref<PredictionWithOdds[]>([])
  const currentPrediction = ref<PredictionWithOdds | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchPredictions() {
    loading.value = true
    error.value = null
    try {
      predictions.value = await api.getPredictions()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load predictions'
    } finally {
      loading.value = false
    }
  }

  async function fetchPrediction(id: string) {
    loading.value = true
    error.value = null
    try {
      currentPrediction.value = await api.getPrediction(id)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load prediction'
      currentPrediction.value = null
    } finally {
      loading.value = false
    }
  }

  function clearCurrent() {
    currentPrediction.value = null
  }

  return {
    predictions,
    currentPrediction,
    loading,
    error,
    fetchPredictions,
    fetchPrediction,
    clearCurrent,
  }
})
