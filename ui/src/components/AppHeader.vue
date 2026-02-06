<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'

const authStore = useAuthStore()
const router = useRouter()

function goToLeaderboard() {
  router.push({ name: 'leaderboard' })
}

// Spin state
const rotation = ref(0)
const velocity = ref(0)
let lastX = 0
let isDragging = false
let animationId: number | null = null
let swipeDistance = 0
let resetTimeout: ReturnType<typeof setTimeout> | null = null
const SPIN_THRESHOLD = 30

function cancelReset() {
  if (resetTimeout !== null) {
    clearTimeout(resetTimeout)
    resetTimeout = null
  }
}

function scheduleReset() {
  cancelReset()
  resetTimeout = setTimeout(() => {
    resetTimeout = null
    startResetAnimation()
  }, 3000)
}

function startResetAnimation() {
  // Normalize rotation to [-180, 180] range so it takes the shortest path
  let target = 0
  let current = rotation.value % 360
  if (current > 180) current -= 360
  if (current < -180) current += 360
  rotation.value = current

  function tick() {
    rotation.value *= 0.85
    if (Math.abs(rotation.value) > 0.5) {
      animationId = requestAnimationFrame(tick)
    } else {
      rotation.value = 0
      animationId = null
    }
  }
  animationId = requestAnimationFrame(tick)
}

function onPointerDown(e: PointerEvent) {
  isDragging = true
  lastX = e.clientX
  swipeDistance = 0
  velocity.value = 0
  stopMomentum()
  cancelReset()
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!isDragging) return
  const dx = e.clientX - lastX
  lastX = e.clientX
  const delta = dx * 3
  rotation.value += delta
  velocity.value = delta
  swipeDistance += Math.abs(dx)
}

function onPointerUp() {
  if (!isDragging) return
  isDragging = false
  if (swipeDistance >= SPIN_THRESHOLD) {
    api.spin().catch(() => {})
  }
  startMomentum()
}

function startMomentum() {
  function tick() {
    velocity.value *= 0.95
    rotation.value += velocity.value

    if (Math.abs(velocity.value) > 0.1) {
      animationId = requestAnimationFrame(tick)
    } else {
      velocity.value = 0
      animationId = null
      scheduleReset()
    }
  }
  animationId = requestAnimationFrame(tick)
}

function stopMomentum() {
  if (animationId !== null) {
    cancelAnimationFrame(animationId)
    animationId = null
  }
}

onUnmounted(() => {
  stopMomentum()
  cancelReset()
})
</script>

<template>
  <header class="bg-dark px-4 py-3 flex items-center justify-between border-b border-dark-lighter">
    <div class="flex items-center gap-3">
      <div class="avatar-perspective">
        <div
          class="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-secondary flex items-center justify-center font-bold text-dark select-none touch-none cursor-grab active:cursor-grabbing"
          :style="{ transform: `rotateY(${rotation}deg)` }"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
        >
          {{ authStore.user?.name.charAt(0).toUpperCase() }}
        </div>
      </div>
      <div>
        <p class="font-medium">{{ authStore.user?.name }}</p>
        <p class="text-sm text-gray-400">Player</p>
      </div>
    </div>
    <button
      @click="goToLeaderboard"
      class="bg-dark-light hover:bg-dark-lighter px-4 py-2 rounded-full flex items-center gap-2 transition-colors"
    >
      <span class="text-primary font-bold">{{ authStore.user?.tokens ?? 0 }}</span>
      <span class="text-gray-400 text-sm">tokens</span>
    </button>
  </header>
</template>

<style scoped>
.avatar-perspective {
  perspective: 200px;
}
</style>
