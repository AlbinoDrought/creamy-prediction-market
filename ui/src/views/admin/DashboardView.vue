<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { PredictionWithOdds } from '@/types/predictions'
import { PredictionStatus } from '@/types/predictions'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import Toast from '@/components/Toast.vue'

const router = useRouter()
const predictions = ref<PredictionWithOdds[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const showToast = ref(false)
const actionLoading = ref<string | null>(null)

onMounted(async () => {
  await fetchPredictions()
})

async function fetchPredictions() {
  loading.value = true
  try {
    predictions.value = await api.getPredictions()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load predictions'
  } finally {
    loading.value = false
  }
}

function goToCreate() {
  router.push({ name: 'admin-prediction-new' })
}

function goToEdit(id: string) {
  router.push({ name: 'admin-prediction-edit', params: { id } })
}

async function closePrediction(id: string) {
  actionLoading.value = id
  try {
    await api.closePrediction(id)
    await fetchPredictions()
    toastType.value = 'success'
    toastMessage.value = 'Prediction closed'
    showToast.value = true
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to close prediction'
    showToast.value = true
  } finally {
    actionLoading.value = null
  }
}

async function reopenPrediction(id: string) {
  actionLoading.value = id
  try {
    await api.reopenPrediction(id)
    await fetchPredictions()
    toastType.value = 'success'
    toastMessage.value = 'Prediction re-opened'
    showToast.value = true
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to re-open prediction'
    showToast.value = true
  } finally {
    actionLoading.value = null
  }
}

async function voidPrediction(id: string) {
  if (!confirm('Are you sure you want to void this prediction? All bets will be refunded.')) return

  actionLoading.value = id
  try {
    await api.voidPrediction(id)
    await fetchPredictions()
    toastType.value = 'success'
    toastMessage.value = 'Prediction voided'
    showToast.value = true
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to void prediction'
    showToast.value = true
  } finally {
    actionLoading.value = null
  }
}

const decideModal = ref<{ id: string; choices: PredictionWithOdds['prediction']['choices'] } | null>(null)

function openDecideModal(prediction: PredictionWithOdds) {
  decideModal.value = { id: prediction.prediction.id, choices: prediction.prediction.choices }
}

async function decidePrediction(winningChoiceId: string) {
  if (!decideModal.value) return

  actionLoading.value = decideModal.value.id
  try {
    await api.decidePrediction(decideModal.value.id, winningChoiceId)
    decideModal.value = null
    await fetchPredictions()
    toastType.value = 'success'
    toastMessage.value = 'Winner decided!'
    showToast.value = true
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to decide prediction'
    showToast.value = true
  } finally {
    actionLoading.value = null
  }
}

function statusClass(status: string) {
  switch (status) {
    case PredictionStatus.Open:
      return 'bg-success/20 text-success'
    case PredictionStatus.Closed:
      return 'bg-warning/20 text-warning'
    case PredictionStatus.Decided:
      return 'bg-secondary/20 text-secondary-light'
    case PredictionStatus.Void:
      return 'bg-gray-500/20 text-gray-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-xl font-bold text-white">Predictions</h2>
      <button
        @click="goToCreate"
        class="bg-secondary hover:bg-secondary-dark text-white px-4 py-2 rounded-lg font-medium transition-colors"
      >
        + New Prediction
      </button>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>

    <div v-else-if="error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error">
      {{ error }}
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="prediction in predictions"
        :key="prediction.prediction.id"
        class="bg-dark-light rounded-xl p-4"
      >
        <div class="flex items-start justify-between gap-3 mb-3">
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-white">{{ prediction.prediction.name }}</h3>
            <p v-if="prediction.prediction.description" class="text-sm text-gray-400 truncate">
              {{ prediction.prediction.description }}
            </p>
          </div>
          <span
            class="px-2 py-1 rounded-full text-xs font-medium shrink-0"
            :class="statusClass(prediction.prediction.status)"
          >
            {{ prediction.prediction.status }}
          </span>
        </div>

        <div class="flex flex-wrap gap-2 mb-4">
          <span
            v-for="choice in prediction.prediction.choices"
            :key="choice.id"
            class="px-2 py-1 rounded text-xs"
            :class="prediction.prediction.winning_choice_id === choice.id
              ? 'bg-success/20 text-success font-medium'
              : 'bg-dark text-gray-300'"
          >
            {{ choice.name }}
            <template v-if="prediction.prediction.winning_choice_id === choice.id"> (Winner)</template>
          </span>
        </div>

        <div class="flex flex-wrap gap-2">
          <button
            v-if="prediction.prediction.status === PredictionStatus.Open"
            @click="goToEdit(prediction.prediction.id)"
            class="bg-dark hover:bg-dark-lighter text-white px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            Edit
          </button>
          <button
            v-if="prediction.prediction.status === PredictionStatus.Open"
            @click="closePrediction(prediction.prediction.id)"
            :disabled="actionLoading === prediction.prediction.id"
            class="bg-warning/20 hover:bg-warning/30 text-warning px-3 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            {{ actionLoading === prediction.prediction.id ? 'Closing...' : 'Close Betting' }}
          </button>
          <button
            v-if="prediction.prediction.status === PredictionStatus.Closed"
            @click="reopenPrediction(prediction.prediction.id)"
            :disabled="actionLoading === prediction.prediction.id"
            class="bg-secondary/20 hover:bg-secondary/30 text-secondary-light px-3 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            {{ actionLoading === prediction.prediction.id ? 'Re-opening...' : 'Re-open Betting' }}
          </button>
          <button
            v-if="prediction.prediction.status === PredictionStatus.Closed"
            @click="openDecideModal(prediction)"
            :disabled="actionLoading === prediction.prediction.id"
            class="bg-success/20 hover:bg-success/30 text-success px-3 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            Decide Winner
          </button>
          <button
            v-if="prediction.prediction.status === PredictionStatus.Open || prediction.prediction.status === PredictionStatus.Closed"
            @click="voidPrediction(prediction.prediction.id)"
            :disabled="actionLoading === prediction.prediction.id"
            class="bg-error/20 hover:bg-error/30 text-error px-3 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            Void
          </button>
        </div>
      </div>

      <div v-if="predictions.length === 0" class="text-center py-12">
        <p class="text-gray-400 mb-4">No predictions yet</p>
        <button
          @click="goToCreate"
          class="bg-secondary text-white px-4 py-2 rounded-lg font-medium"
        >
          Create First Prediction
        </button>
      </div>
    </div>

    <!-- Decide Modal -->
    <Teleport to="body">
      <div
        v-if="decideModal"
        class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50"
        @click.self="decideModal = null"
      >
        <div class="bg-dark rounded-xl p-6 w-full max-w-sm">
          <h3 class="text-lg font-bold text-white mb-4">Select Winner</h3>
          <div class="space-y-2 mb-4">
            <button
              v-for="choice in decideModal.choices"
              :key="choice.id"
              @click="decidePrediction(choice.id)"
              :disabled="actionLoading !== null"
              class="w-full bg-dark-light hover:bg-dark-lighter text-white px-4 py-3 rounded-lg font-medium transition-colors text-left disabled:opacity-50"
            >
              {{ choice.name }}
            </button>
          </div>
          <button
            @click="decideModal = null"
            class="w-full text-gray-400 hover:text-white py-2"
          >
            Cancel
          </button>
        </div>
      </div>
    </Teleport>

    <Toast
      v-if="showToast"
      :message="toastMessage"
      :type="toastType"
      @close="showToast = false"
    />
  </div>
</template>
