<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import type { Prediction } from '@/types/predictions'
import { PredictionStatus } from '@/types/predictions'

const props = defineProps<{
  prediction: Prediction
}>()

const now = ref(new Date())
let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  timer = setInterval(() => {
    now.value = new Date()
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const statusConfig = computed(() => {
  switch (props.prediction.status) {
    case PredictionStatus.Open:
      return { label: 'Open', class: 'bg-success/20 text-success' }
    case PredictionStatus.Closed:
      return { label: 'Closed', class: 'bg-warning/20 text-warning' }
    case PredictionStatus.Decided:
      return { label: 'Decided', class: 'bg-secondary/20 text-secondary-light' }
    case PredictionStatus.Void:
      return { label: 'Void', class: 'bg-gray-500/20 text-gray-400' }
    default:
      return { label: 'Unknown', class: 'bg-gray-500/20 text-gray-400' }
  }
})

const timeRemaining = computed(() => {
  if (props.prediction.status !== PredictionStatus.Open) return null

  const closes = new Date(props.prediction.closes_at)
  const diff = closes.getTime() - now.value.getTime()

  if (diff <= 0) return 'Closing soon...'

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diff % (1000 * 60)) / 1000)

  if (hours > 24) {
    const days = Math.floor(hours / 24)
    return `${days}d ${hours % 24}h`
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m ${seconds}s`
})
</script>

<template>
  <router-link
    :to="{ name: 'prediction', params: { id: prediction.id } }"
    class="block bg-dark-light rounded-xl p-4 hover:bg-dark-lighter transition-colors"
  >
    <div class="flex items-start justify-between gap-3 mb-3">
      <h3 class="font-semibold text-white flex-1">{{ prediction.name }}</h3>
      <span
        class="px-2 py-1 rounded-full text-xs font-medium shrink-0"
        :class="statusConfig.class"
      >
        {{ statusConfig.label }}
      </span>
    </div>

    <p v-if="prediction.description" class="text-gray-400 text-sm mb-3 line-clamp-2">
      {{ prediction.description }}
    </p>

    <div class="flex items-center justify-between">
      <div class="flex gap-2">
        <span
          v-for="choice in prediction.choices.slice(0, 2)"
          :key="choice.id"
          class="bg-dark px-2 py-1 rounded text-xs text-gray-300"
        >
          {{ choice.name }}
        </span>
        <span
          v-if="prediction.choices.length > 2"
          class="bg-dark px-2 py-1 rounded text-xs text-gray-400"
        >
          +{{ prediction.choices.length - 2 }} more
        </span>
      </div>

      <div v-if="timeRemaining" class="text-xs text-primary font-medium">
        {{ timeRemaining }}
      </div>
    </div>
  </router-link>
</template>
