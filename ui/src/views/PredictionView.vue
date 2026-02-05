<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePredictionsStore } from '@/stores/predictions'
import { useBetsStore } from '@/stores/bets'
import { useAuthStore } from '@/stores/auth'
import { PredictionStatus } from '@/types/predictions'
import AppHeader from '@/components/AppHeader.vue'
import BottomNav from '@/components/BottomNav.vue'
import ChoiceButton from '@/components/ChoiceButton.vue'
import BetAmountInput from '@/components/BetAmountInput.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import Toast from '@/components/Toast.vue'

const route = useRoute()
const router = useRouter()
const predictionsStore = usePredictionsStore()
const betsStore = useBetsStore()
const authStore = useAuthStore()

const selectedChoice = ref<string | null>(null)
const betAmount = ref(10)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const showToast = ref(false)

const predictionId = computed(() => route.params.id as string)

onMounted(async () => {
  await predictionsStore.fetchPrediction(predictionId.value)
  await betsStore.fetchBets()
})

watch(predictionId, async (newId) => {
  await predictionsStore.fetchPrediction(newId)
})

const prediction = computed(() => predictionsStore.currentPrediction?.prediction)
const odds = computed(() => predictionsStore.currentPrediction?.odds)

const userBets = computed(() => {
  if (!prediction.value) return []
  return betsStore.getBetsForPrediction(prediction.value.id)
})

const canBet = computed(() => {
  return prediction.value?.status === PredictionStatus.Open &&
    authStore.user &&
    authStore.user.tokens > 0
})

const showOdds = computed(() => {
  if (!prediction.value) return true
  // Show odds if the prediction allows it, or if user has already bet
  return prediction.value.odds_visible_before_bet || userBets.value.length > 0
})

function getOddsForChoice(choiceId: string) {
  return odds.value?.choices.find(c => c.prediction_choice_id === choiceId)
}

async function placeBet() {
  if (!selectedChoice.value || !canBet.value) return

  try {
    await betsStore.placeBet(predictionId.value, selectedChoice.value, betAmount.value)
    toastType.value = 'success'
    toastMessage.value = `Bet placed! ${betAmount.value} tokens on your choice`
    showToast.value = true
    selectedChoice.value = null
    betAmount.value = Math.min(10, authStore.user?.tokens ?? 10)
    // Refresh prediction to get updated odds
    await predictionsStore.fetchPrediction(predictionId.value)
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to place bet'
    showToast.value = true
  }
}

function goBack() {
  router.push({ name: 'home' })
}
</script>

<template>
  <div class="min-h-screen bg-bg pb-20">
    <AppHeader />

    <main class="p-4">
      <!-- Back button -->
      <button @click="goBack" class="flex items-center gap-2 text-gray-400 hover:text-white mb-4">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back
      </button>

      <div v-if="predictionsStore.loading" class="flex justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>

      <div v-else-if="predictionsStore.error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error">
        {{ predictionsStore.error }}
      </div>

      <template v-else-if="prediction">
        <!-- Prediction Info -->
        <div class="mb-6">
          <div class="flex items-start justify-between gap-3 mb-2">
            <h1 class="text-2xl font-bold text-white">{{ prediction.name }}</h1>
            <span
              class="px-3 py-1 rounded-full text-sm font-medium shrink-0"
              :class="{
                'bg-success/20 text-success': prediction.status === PredictionStatus.Open,
                'bg-warning/20 text-warning': prediction.status === PredictionStatus.Closed,
                'bg-secondary/20 text-secondary-light': prediction.status === PredictionStatus.Decided,
                'bg-gray-500/20 text-gray-400': prediction.status === PredictionStatus.Void,
              }"
            >
              {{ prediction.status }}
            </span>
          </div>
          <p v-if="prediction.description" class="text-gray-400">
            {{ prediction.description }}
          </p>
        </div>

        <!-- User's existing bets -->
        <div v-if="userBets.length > 0" class="mb-6">
          <h2 class="text-lg font-semibold text-white mb-3">Your Bets</h2>
          <div class="space-y-2">
            <div
              v-for="bet in userBets"
              :key="bet.id"
              class="bg-dark-light rounded-lg p-3 flex items-center justify-between"
            >
              <div>
                <span class="text-white font-medium">
                  {{ prediction.choices.find(c => c.id === bet.prediction_choice_id)?.name }}
                </span>
                <span class="text-gray-400 ml-2">{{ bet.amount }} tokens</span>
              </div>
              <span
                class="px-2 py-1 rounded text-xs font-medium"
                :class="{
                  'bg-primary/20 text-primary': bet.status === 'placed',
                  'bg-success/20 text-success': bet.status === 'won',
                  'bg-error/20 text-error': bet.status === 'lost',
                  'bg-gray-500/20 text-gray-400': bet.status === 'voided',
                }"
              >
                {{ bet.status }}
                <template v-if="bet.status === 'won'"> (+{{ bet.won_amount }})</template>
              </span>
            </div>
          </div>
        </div>

        <!-- Betting Section -->
        <div v-if="canBet" class="space-y-6">
          <div>
            <h2 class="text-lg font-semibold text-white mb-3">Make Your Pick</h2>
            <div class="space-y-3">
              <ChoiceButton
                v-for="choice in prediction.choices"
                :key="choice.id"
                :choice="choice"
                :odds="getOddsForChoice(choice.id)"
                :selected="selectedChoice === choice.id"
                :disabled="false"
                :show-odds="showOdds"
                @select="selectedChoice = $event"
              />
            </div>
          </div>

          <Transition name="fade">
            <div v-if="selectedChoice" class="space-y-4">
              <h2 class="text-lg font-semibold text-white">Bet Amount</h2>
              <BetAmountInput
                v-model="betAmount"
                :max="authStore.user?.tokens ?? 0"
              />

              <button
                @click="placeBet"
                :disabled="betsStore.placingBet || !selectedChoice"
                class="w-full bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-dark font-bold py-4 px-4 rounded-xl transition-colors flex items-center justify-center gap-2"
              >
                <LoadingSpinner v-if="betsStore.placingBet" size="sm" />
                <span v-else>Place Bet ({{ betAmount }} tokens)</span>
              </button>
            </div>
          </Transition>
        </div>

        <!-- View only mode for closed/decided predictions -->
        <div v-else-if="prediction.status !== PredictionStatus.Open">
          <h2 class="text-lg font-semibold text-white mb-3">Choices</h2>
          <div class="space-y-3">
            <ChoiceButton
              v-for="choice in prediction.choices"
              :key="choice.id"
              :choice="choice"
              :odds="getOddsForChoice(choice.id)"
              :selected="false"
              :disabled="true"
              :show-odds="true"
              @select="() => {}"
            />
          </div>
        </div>

        <!-- No tokens warning -->
        <div v-else-if="authStore.user?.tokens === 0" class="bg-warning/20 border border-warning rounded-lg px-4 py-3 text-warning text-center">
          You're out of tokens! Check the leaderboard to see how you're doing.
        </div>
      </template>
    </main>

    <BottomNav />

    <Toast
      v-if="showToast"
      :message="toastMessage"
      :type="toastType"
      @close="showToast = false"
    />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
