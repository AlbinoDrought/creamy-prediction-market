<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: number
  max: number
  min?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const minValue = computed(() => props.min ?? 1)

function updateValue(value: number) {
  const clamped = Math.max(minValue.value, Math.min(props.max, Math.round(value)))
  emit('update:modelValue', clamped)
}

const quickAmounts = computed(() => {
  const amounts = [10, 25, 50, 100]
  return amounts.filter(a => a <= props.max)
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-4">
      <input
        type="range"
        :min="minValue"
        :max="max"
        :value="modelValue"
        @input="updateValue(Number(($event.target as HTMLInputElement).value))"
        class="flex-1 h-2 bg-dark rounded-lg appearance-none cursor-pointer accent-primary"
      />
      <div class="relative">
        <input
          type="number"
          :min="minValue"
          :max="max"
          :value="modelValue"
          @input="updateValue(Number(($event.target as HTMLInputElement).value))"
          class="w-24 bg-dark border border-dark-lighter rounded-lg px-3 py-2 text-white text-center focus:outline-none focus:border-primary"
        />
      </div>
    </div>

    <div class="flex gap-2">
      <button
        v-for="amount in quickAmounts"
        :key="amount"
        @click="updateValue(amount)"
        class="flex-1 py-2 px-3 rounded-lg text-sm font-medium transition-colors"
        :class="modelValue === amount
          ? 'bg-primary text-dark'
          : 'bg-dark-light text-gray-300 hover:bg-dark-lighter'"
      >
        {{ amount }}
      </button>
      <button
        @click="updateValue(max)"
        class="flex-1 py-2 px-3 rounded-lg text-sm font-medium transition-colors"
        :class="modelValue === max
          ? 'bg-primary text-dark'
          : 'bg-dark-light text-gray-300 hover:bg-dark-lighter'"
      >
        Max
      </button>
    </div>

    <div class="text-center text-sm text-gray-400">
      Available: <span class="text-primary font-medium">{{ max }}</span> tokens
    </div>
  </div>
</template>
