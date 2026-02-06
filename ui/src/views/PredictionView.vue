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
const showIncreaseBetUI = ref(false)
const betAmount = ref(10)
const increaseAmount = ref(1)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const showToast = ref(false)

// Swipe-to-dismiss state (shared for both footers)
const touchStartY = ref(0)
const touchCurrentY = ref(0)
const isDragging = ref(false)
const swipeDismissed = ref(false)
const swipeDismissedIncrease = ref(false)
const SWIPE_THRESHOLD = 80

function onTouchStart(e: TouchEvent) {
  if (!e.touches[0]) return
  touchStartY.value = e.touches[0].clientY
  touchCurrentY.value = e.touches[0].clientY
  isDragging.value = true
}

function onTouchMove(e: TouchEvent) {
  if (!isDragging.value) return
  if (!e.touches[0]) return
  touchCurrentY.value = e.touches[0].clientY
}

function onTouchEnd() {
  if (!isDragging.value) return
  const swipeDistance = touchCurrentY.value - touchStartY.value
  if (swipeDistance > SWIPE_THRESHOLD) {
    // Swiped down far enough - dismiss immediately (skip Vue transition)
    swipeDismissed.value = true
    selectedChoice.value = null
    setTimeout(() => {
      swipeDismissed.value = false
    }, 50)
  }
  isDragging.value = false
  touchStartY.value = 0
  touchCurrentY.value = 0
}

function onTouchEndIncrease() {
  if (!isDragging.value) return
  const swipeDistance = touchCurrentY.value - touchStartY.value
  if (swipeDistance > SWIPE_THRESHOLD) {
    swipeDismissedIncrease.value = true
    showIncreaseBetUI.value = false
    setTimeout(() => {
      swipeDismissedIncrease.value = false
    }, 50)
  }
  isDragging.value = false
  touchStartY.value = 0
  touchCurrentY.value = 0
}

const swipeOffset = computed(() => {
  if (!isDragging.value) return 0
  const offset = touchCurrentY.value - touchStartY.value
  // Only allow dragging down, with resistance
  return offset > 0 ? Math.min(offset * 0.6, 150) : 0
})

// Auto-scroll selected choice into view
function onChoiceSelect(choiceId: string) {
  const isFirstSelection = !selectedChoice.value
  selectedChoice.value = choiceId
  // Cap bet amount to available tokens
  const maxTokens = authStore.user?.tokens ?? 0
  if (betAmount.value > maxTokens) {
    betAmount.value = Math.max(1, maxTokens)
  }
  // Only scroll when footer first opens, not when switching choices
  if (isFirstSelection) {
    setTimeout(() => {
      const el = document.getElementById(`choice-${choiceId}`)
      if (el) {
        // Scroll so element is in upper third of visible area (above footer)
        const rect = el.getBoundingClientRect()
        const targetY = window.scrollY + rect.top - window.innerHeight / 3
        window.scrollTo({ top: targetY, behavior: 'smooth' })
      }
    }, 100)
  }
}

function openIncreaseBetUI() {
  // Reset increase amount to minimum (1 token)
  increaseAmount.value = 1
  showIncreaseBetUI.value = true
}

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

// Get the user's existing bet for this prediction (only one allowed)
const existingBet = computed(() => {
  if (!prediction.value) return null
  return betsStore.getBetForPrediction(prediction.value.id)
})

// Can place a NEW bet (only if no existing bet)
const canPlaceNewBet = computed(() => {
  return prediction.value?.status === PredictionStatus.Open &&
    authStore.user &&
    authStore.user.tokens > 0 &&
    !existingBet.value
})

// Can increase an existing bet
const canIncreaseBet = computed(() => {
  return prediction.value?.status === PredictionStatus.Open &&
    authStore.user &&
    authStore.user.tokens > 0 &&
    existingBet.value &&
    existingBet.value.status === 'placed'
})

const showOdds = computed(() => {
  if (!prediction.value) return true
  // Show odds if the prediction allows it, or if user has already bet
  return prediction.value.odds_visible_before_bet || !!existingBet.value
})

function getOddsForChoice(choiceId: string) {
  return odds.value?.choices.find(c => c.prediction_choice_id === choiceId)
}

async function placeBet() {
  if (!selectedChoice.value || !canPlaceNewBet.value) return

  try {
    await betsStore.placeBet(predictionId.value, selectedChoice.value, betAmount.value)
    toastType.value = 'success'
    toastMessage.value = `Bet placed! ${betAmount.value} tokens on your choice`
    showToast.value = true
    // Skip transition on dismiss
    swipeDismissed.value = true
    selectedChoice.value = null
    setTimeout(() => { swipeDismissed.value = false }, 50)
    betAmount.value = Math.min(10, authStore.user?.tokens ?? 10)
    // Refresh prediction to get updated odds
    await predictionsStore.fetchPrediction(predictionId.value)
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to place bet'
    showToast.value = true
  }
}

async function increaseBet() {
  if (!existingBet.value || !canIncreaseBet.value) return
  if (!showIncreaseBetUI.value) return

  try {
    const newTotal = existingBet.value.amount + increaseAmount.value
    await betsStore.increaseBet(existingBet.value.id, newTotal)
    toastType.value = 'success'
    toastMessage.value = `Bet increased by ${increaseAmount.value} tokens!`
    showToast.value = true
    // Skip transition on dismiss
    swipeDismissedIncrease.value = true
    showIncreaseBetUI.value = false
    setTimeout(() => { swipeDismissedIncrease.value = false }, 50)
    // Refresh prediction to get updated odds
    await predictionsStore.fetchPrediction(predictionId.value)
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to increase bet'
    showToast.value = true
  }
}


function goBack() {
  router.push({ name: 'home' })
}
</script>

<template>
  <div class="min-h-screen bg-bg" :class="(canPlaceNewBet && selectedChoice) || (canIncreaseBet && showIncreaseBetUI) ? 'pb-72' : 'pb-20'">
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

        <!-- User has already placed a bet -->
        <div v-if="existingBet" class="mb-6">
          <h2 class="text-lg font-semibold text-white mb-3">Your Bet</h2>
          <div class="bg-primary/10 border border-primary/30 rounded-xl p-4">
            <div class="flex items-center justify-between mb-2">
              <div>
                <span class="text-primary font-bold text-lg">
                  {{ prediction.choices.find(c => c.id === existingBet?.prediction_choice_id)?.name }}
                </span>
              </div>
              <span
                class="px-2 py-1 rounded text-xs font-medium"
                :class="{
                  'bg-primary/20 text-primary': existingBet?.status === 'placed',
                  'bg-success/20 text-success': existingBet?.status === 'won',
                  'bg-error/20 text-error': existingBet?.status === 'lost',
                  'bg-gray-500/20 text-gray-400': existingBet?.status === 'voided',
                }"
              >
                {{ existingBet?.status }}
                <template v-if="existingBet?.status === 'won'"> (+{{ existingBet?.won_amount }})</template>
              </span>
            </div>
            <div class="text-gray-400">
              <span class="text-white font-medium">{{ existingBet.amount }}</span> tokens wagered
              <template v-if="existingBet.status === 'placed' && getOddsForChoice(existingBet.prediction_choice_id)">
                <span class="mx-2">→</span>
                <span class="text-success font-medium">
                  {{ Math.floor((existingBet.amount * getOddsForChoice(existingBet.prediction_choice_id)!.odds_basis_points) / 100) }}
                </span> if correct
              </template>
            </div>
          </div>

          <!-- Show all choices with odds (read-only) -->
          <div class="mt-4">
            <h3 class="text-sm font-medium text-gray-400 mb-2">Current Odds</h3>
            <div class="space-y-2">
              <div
                v-for="choice in prediction.choices"
                :key="choice.id"
                class="flex items-center justify-between p-3 rounded-lg"
                :class="choice.id === existingBet.prediction_choice_id ? 'bg-primary/10' : 'bg-dark-light'"
              >
                <span :class="choice.id === existingBet.prediction_choice_id ? 'text-primary font-medium' : 'text-white'">
                  {{ choice.name }}
                  <span v-if="choice.id === existingBet.prediction_choice_id" class="text-xs text-primary/70 ml-1">(your pick)</span>
                </span>
                <div v-if="showOdds && getOddsForChoice(choice.id)" class="text-right">
                  <span class="text-primary font-bold">
                    {{ (getOddsForChoice(choice.id)!.odds_basis_points / 100).toFixed(1) }}x
                  </span>
                  <span class="text-xs text-gray-400 ml-2">
                    {{ getOddsForChoice(choice.id)!.tokens_placed }} tokens
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Increase bet trigger button -->
          <div v-if="canIncreaseBet" class="mt-4">
            <button
              @click="openIncreaseBetUI"
              class="w-full bg-dark-lighter hover:bg-dark-light border border-dark-lighter text-white font-medium py-3 px-4 rounded-xl transition-colors"
            >
              Increase Bet
            </button>
          </div>

          <!-- No more tokens to add -->
          <div v-else-if="existingBet.status === 'placed' && authStore.user?.tokens === 0" class="mt-4 bg-dark-light rounded-lg p-3 text-center text-gray-400 text-sm">
            No more tokens to add to your bet
          </div>
        </div>

        <!-- New bet section (only if no existing bet) -->
        <div v-else-if="canPlaceNewBet">
          <div>
            <h2 class="text-lg font-semibold text-white mb-3">Make Your Pick</h2>
            <div class="space-y-3">
              <ChoiceButton
                v-for="choice in prediction.choices"
                :id="`choice-${choice.id}`"
                :key="choice.id"
                :choice="choice"
                :odds="getOddsForChoice(choice.id)"
                :selected="selectedChoice === choice.id"
                :disabled="false"
                :show-odds="showOdds"
                @select="onChoiceSelect($event)"
              />
            </div>
          </div>
          <!-- Spacer to allow scrolling when footer is visible -->
          <div v-if="selectedChoice" class="h-72"></div>
        </div>

        <!-- View only mode for closed/decided predictions (no existing bet) -->
        <div v-else-if="prediction.status !== PredictionStatus.Open && !existingBet">
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

        <!-- No tokens warning (and no existing bet) -->
        <div v-else-if="authStore.user?.tokens === 0 && !existingBet" class="bg-warning/20 border border-warning rounded-lg px-4 py-3 text-warning text-center">
          You're out of tokens! Check the leaderboard to see how you're doing.
        </div>
      </template>
    </main>

    <!-- Sticky bet action footer for new bets -->
    <Transition :name="swipeDismissed ? 'none' : 'slide-up'">
      <div
        v-if="canPlaceNewBet && selectedChoice"
        class="fixed bottom-0 left-0 right-0 bg-dark-light border-t border-dark-lighter p-4 pb-24 z-10 touch-pan-x"
        :class="{ 'transition-transform duration-200': !isDragging }"
        :style="{ transform: `translateY(${swipeOffset}px)`, opacity: isDragging ? 1 - swipeOffset / 200 : 1 }"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEnd"
      >
        <!-- Drag handle -->
        <div class="flex justify-center mb-3 -mt-1">
          <div class="w-10 h-1 bg-gray-600 rounded-full"></div>
        </div>
        <div class="max-w-md mx-auto space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-white">Bet Amount</h2>
            <span class="text-xs text-gray-400">{{ authStore.user?.tokens ?? 0 }} available</span>
          </div>
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
      </div>
    </Transition>

    <!-- Sticky footer for increasing bet -->
    <Transition :name="swipeDismissedIncrease ? 'none' : 'slide-up'">
      <div
        v-if="canIncreaseBet && showIncreaseBetUI"
        class="fixed bottom-0 left-0 right-0 bg-dark-light border-t border-dark-lighter p-4 pb-24 z-10 touch-pan-x"
        :class="{ 'transition-transform duration-200': !isDragging }"
        :style="{ transform: `translateY(${swipeOffset}px)`, opacity: isDragging ? 1 - swipeOffset / 200 : 1 }"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEndIncrease"
      >
        <!-- Drag handle -->
        <div class="flex justify-center mb-3 -mt-1">
          <div class="w-10 h-1 bg-gray-600 rounded-full"></div>
        </div>
        <div class="max-w-md mx-auto space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-white">Add Tokens</h2>
            <span class="text-xs text-gray-400">{{ authStore.user?.tokens ?? 0 }} available</span>
          </div>
          <BetAmountInput
            v-model="increaseAmount"
            :max="authStore.user?.tokens ?? 0"
          />
          <p class="text-sm text-gray-400 text-center">
            {{ existingBet!.amount }} → <span class="text-primary font-medium">{{ existingBet!.amount + increaseAmount }}</span> tokens
          </p>
          <button
            @click="increaseBet"
            :disabled="betsStore.placingBet || increaseAmount < 1"
            class="w-full bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-dark font-bold py-4 px-4 rounded-xl transition-colors flex items-center justify-center gap-2"
          >
            <LoadingSpinner v-if="betsStore.placingBet" size="sm" />
            <span v-else>Add {{ increaseAmount }} {{ increaseAmount === 1 ? 'token' : 'tokens' }}</span>
          </button>
        </div>
      </div>
    </Transition>

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

.none-enter-active,
.none-leave-active {
  transition: all 0s;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(100%);
}
</style>
