<script setup lang="ts">
import { computed } from 'vue'
import type { PredictionChoice, PredictionChoiceOdds } from '@/types/predictions'

const props = defineProps<{
  choice: PredictionChoice
  odds?: PredictionChoiceOdds
  selected: boolean
  disabled: boolean
  showOdds: boolean
  winner?: boolean
}>()

const emit = defineEmits<{
  select: [choiceId: string]
}>()

const oddsDisplay = computed(() => {
  if (!props.odds || props.odds.odds_basis_points === 0) return null
  // Convert basis points to multiplier (e.g., 250 -> 2.5x)
  return (props.odds.odds_basis_points / 100).toFixed(1) + 'x'
})

const percentage = computed(() => {
  if (!props.odds || props.odds.tokens_placed === 0) return null
  // This would need total tokens from parent, but we can show tokens placed
  return props.odds.tokens_placed
})
</script>

<template>
  <button
    @click="emit('select', choice.id)"
    :disabled="disabled"
    class="w-full p-4 rounded-xl border-2 transition-all text-left"
    :class="[
      winner
        ? 'border-success bg-success/10'
        : selected
          ? 'border-primary bg-primary/10'
          : 'border-dark-lighter bg-dark-light hover:border-dark-lighter hover:bg-dark-lighter',
      disabled && !winner && 'opacity-50 cursor-not-allowed'
    ]"
  >
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div
          class="w-5 h-5 rounded-full border-2 flex items-center justify-center"
          :class="selected ? 'border-primary' : 'border-gray-500'"
        >
          <div v-if="selected" class="w-3 h-3 rounded-full bg-primary" />
        </div>
        <span class="font-medium" :class="winner ? 'text-success' : selected ? 'text-primary' : 'text-white'">
          {{ choice.name }}
        </span>
        <span v-if="winner" class="ml-2 text-xs font-medium text-success">Winner</span>
      </div>

      <div v-if="showOdds && oddsDisplay" class="text-right">
        <div class="text-primary font-bold">{{ oddsDisplay }}</div>
        <div class="text-xs text-gray-400">{{ percentage }} tokens</div>
      </div>
    </div>
  </button>
</template>
