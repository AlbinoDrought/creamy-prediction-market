<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import Toast from '@/components/Toast.vue'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)
const predictionId = computed(() => route.params.id as string | undefined)

const loading = ref(false)
const saving = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const showToast = ref(false)

const form = ref({
  name: '',
  description: '',
  closes_at: '',
  choices: [{ name: '' }, { name: '' }],
  odds_visible_before_bet: true,
})

onMounted(async () => {
  if (isEdit.value && predictionId.value) {
    loading.value = true
    try {
      const { prediction } = await api.getPrediction(predictionId.value)
      form.value = {
        name: prediction.name,
        description: prediction.description,
        closes_at: prediction.closes_at.slice(0, 16), // Format for datetime-local
        choices: prediction.choices.map(c => ({ name: c.name })),
        odds_visible_before_bet: prediction.odds_visible_before_bet,
      }
    } catch (e) {
      toastType.value = 'error'
      toastMessage.value = e instanceof Error ? e.message : 'Failed to load prediction'
      showToast.value = true
    } finally {
      loading.value = false
    }
  } else {
    // Default closes_at to 1 hour from now
    const now = new Date()
    now.setHours(now.getHours() + 1)
    form.value.closes_at = now.toISOString().slice(0, 16)
  }
})

function addChoice() {
  form.value.choices.push({ name: '' })
}

function removeChoice(index: number) {
  if (form.value.choices.length > 2) {
    form.value.choices.splice(index, 1)
  }
}

async function handleSubmit() {
  if (!form.value.name.trim()) {
    toastType.value = 'error'
    toastMessage.value = 'Name is required'
    showToast.value = true
    return
  }

  const validChoices = form.value.choices.filter(c => c.name.trim())
  if (validChoices.length < 2) {
    toastType.value = 'error'
    toastMessage.value = 'At least 2 choices are required'
    showToast.value = true
    return
  }

  saving.value = true
  try {
    if (isEdit.value && predictionId.value) {
      await api.updatePrediction(predictionId.value, {
        name: form.value.name,
        description: form.value.description,
        closes_at: new Date(form.value.closes_at).toISOString(),
        odds_visible_before_bet: form.value.odds_visible_before_bet,
      })
      toastType.value = 'success'
      toastMessage.value = 'Prediction updated'
      showToast.value = true
    } else {
      await api.createPrediction({
        name: form.value.name,
        description: form.value.description,
        closes_at: new Date(form.value.closes_at).toISOString(),
        choices: validChoices,
        odds_visible_before_bet: form.value.odds_visible_before_bet,
      })
      toastType.value = 'success'
      toastMessage.value = 'Prediction created'
      showToast.value = true
      router.push({ name: 'admin-dashboard' })
    }
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to save prediction'
    showToast.value = true
  } finally {
    saving.value = false
  }
}

function goBack() {
  router.push({ name: 'admin-dashboard' })
}
</script>

<template>
  <div>
    <button @click="goBack" class="flex items-center gap-2 text-gray-400 hover:text-white mb-4">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
      Back
    </button>

    <h2 class="text-xl font-bold text-white mb-6">
      {{ isEdit ? 'Edit Prediction' : 'New Prediction' }}
    </h2>

    <div v-if="loading" class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>

    <form v-else @submit.prevent="handleSubmit" class="space-y-6">
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-2">Name *</label>
        <input
          v-model="form.name"
          type="text"
          required
          class="w-full bg-dark border border-dark-lighter rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-secondary"
          placeholder="e.g., Who will win the Superbowl?"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-300 mb-2">Description</label>
        <textarea
          v-model="form.description"
          rows="3"
          class="w-full bg-dark border border-dark-lighter rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-secondary resize-none"
          placeholder="Optional description..."
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-300 mb-2">Closes At *</label>
        <input
          v-model="form.closes_at"
          type="datetime-local"
          required
          class="w-full bg-dark border border-dark-lighter rounded-lg px-4 py-3 text-white focus:outline-none focus:border-secondary"
        />
      </div>

      <div>
        <label class="flex items-center gap-3 cursor-pointer">
          <input
            v-model="form.odds_visible_before_bet"
            type="checkbox"
            class="w-5 h-5 rounded bg-dark border-dark-lighter text-secondary focus:ring-secondary"
          />
          <span class="text-gray-300">Show odds before placing bet</span>
        </label>
        <p class="text-sm text-gray-500 mt-1 ml-8">
          If disabled, users only see odds after placing their first bet
        </p>
      </div>

      <div v-if="!isEdit">
        <label class="block text-sm font-medium text-gray-300 mb-2">Choices *</label>
        <div class="space-y-2">
          <div
            v-for="(choice, index) in form.choices"
            :key="index"
            class="flex gap-2"
          >
            <input
              v-model="choice.name"
              type="text"
              class="flex-1 bg-dark border border-dark-lighter rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-secondary"
              :placeholder="`Choice ${index + 1}`"
            />
            <button
              v-if="form.choices.length > 2"
              type="button"
              @click="removeChoice(index)"
              class="px-3 text-error hover:text-error-light"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        <button
          type="button"
          @click="addChoice"
          class="mt-2 text-secondary hover:text-secondary-light text-sm font-medium"
        >
          + Add Choice
        </button>
      </div>

      <div v-if="isEdit" class="bg-dark-light rounded-lg p-4">
        <p class="text-sm text-gray-400">
          Note: Choices cannot be modified after creation. To change choices, void this prediction and create a new one.
        </p>
      </div>

      <button
        type="submit"
        :disabled="saving"
        class="w-full bg-secondary hover:bg-secondary-dark disabled:opacity-50 text-white font-bold py-3 px-4 rounded-lg transition-colors flex items-center justify-center gap-2"
      >
        <LoadingSpinner v-if="saving" size="sm" />
        <span v-else>{{ isEdit ? 'Update Prediction' : 'Create Prediction' }}</span>
      </button>
    </form>

    <Toast
      v-if="showToast"
      :message="toastMessage"
      :type="toastType"
      @close="showToast = false"
    />
  </div>
</template>
