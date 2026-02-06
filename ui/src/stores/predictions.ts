import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PredictionWithOdds } from '@/types/predictions'
import { api } from '@/api/client'

export const usePredictionsStore = defineStore('predictions', () => {
  const predictions = ref<PredictionWithOdds[]>([])
  const currentPrediction = ref<PredictionWithOdds | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function swapPredictions() {
    error.value = null
    try {
      predictions.value = await api.getPredictions()
      const cur = currentPrediction.value
      if (cur) {
        currentPrediction.value = predictions.value.find(p => p.prediction.id === cur.prediction.id) || null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load predictions'
    }
  }

  async function fetchPredictions() {
    const loadingTimeout = setTimeout(() => {
      loading.value = true
    }, 70)
    try {
      await swapPredictions()
    } finally {
      clearTimeout(loadingTimeout)
      loading.value = false
    }
  }

  async function fetchPrediction(id: string) {
    const loadingTimeout = setTimeout(() => {
      loading.value = true
    }, 70)
    try {
      await swapPredictions()
      currentPrediction.value = predictions.value.find(p => p.prediction.id === id) || null
      if (!currentPrediction.value && !error.value) {
        error.value = 'Failed to load prediction'
      }
    } finally {
      clearTimeout(loadingTimeout)
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
    swapPredictions,
    fetchPredictions,
    fetchPrediction,
    clearCurrent,
  }
})
