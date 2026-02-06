<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import UserAvatar from '@/components/UserAvatar.vue'
import UserName from '@/components/UserName.vue'

const authStore = useAuthStore()
const router = useRouter()

const userCosmetics = computed(() => authStore.user?.cosmetics)
const titleText = computed(() => userCosmetics.value?.title || 'Player')

function goToShop() {
  router.push({ name: 'shop' })
}

function goToLeaderboard() {
  router.push({ name: 'leaderboard' })
}

// Spin state — module-level so it persists across remounts
const spinState = (() => {
  // Singleton: reuse across component instances
  const key = '__avatarSpin'
  const w = window as any as Record<string, unknown>
  if (w[key]) return w[key] as ReturnType<typeof createSpinState>
  const state = createSpinState()
  w[key] = state
  return state
})()

function createSpinState() {
  return {
    rotation: ref(0),
    velocity: ref(0),
    lastX: 0,
    isDragging: false,
    animationId: null as number | null,
    swipeDistance: 0,
    resetTimeout: null as ReturnType<typeof setTimeout> | null,
  }
}

const rotation = spinState.rotation
const SPIN_THRESHOLD = 30

function cancelReset() {
  if (spinState.resetTimeout !== null) {
    clearTimeout(spinState.resetTimeout)
    spinState.resetTimeout = null
  }
}

function scheduleReset() {
  cancelReset()
  spinState.resetTimeout = setTimeout(() => {
    spinState.resetTimeout = null
    startResetAnimation()
  }, 3000)
}

function startResetAnimation() {
  // Normalize rotation to [-180, 180] range so it takes the shortest path
  let current = rotation.value % 360
  if (current > 180) current -= 360
  if (current < -180) current += 360
  rotation.value = current

  function tick() {
    rotation.value *= 0.85
    if (Math.abs(rotation.value) > 0.5) {
      spinState.animationId = requestAnimationFrame(tick)
    } else {
      rotation.value = 0
      spinState.animationId = null
    }
  }
  spinState.animationId = requestAnimationFrame(tick)
}

function onPointerDown(e: PointerEvent) {
  spinState.isDragging = true
  spinState.lastX = e.clientX
  spinState.swipeDistance = 0
  spinState.velocity.value = 0
  stopMomentum()
  cancelReset()
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!spinState.isDragging) return
  const dx = e.clientX - spinState.lastX
  spinState.lastX = e.clientX
  const delta = dx * 3
  rotation.value += delta
  spinState.velocity.value = delta
  spinState.swipeDistance += Math.abs(dx)
}

function onPointerUp() {
  if (!spinState.isDragging) return
  spinState.isDragging = false
  if (spinState.swipeDistance >= SPIN_THRESHOLD) {
    api.spin().catch(() => {})
  }
  startMomentum()
}

function startMomentum() {
  function tick() {
    spinState.velocity.value *= 0.95
    rotation.value += spinState.velocity.value

    if (Math.abs(spinState.velocity.value) > 0.1) {
      spinState.animationId = requestAnimationFrame(tick)
    } else {
      spinState.velocity.value = 0
      spinState.animationId = null
      scheduleReset()
    }
  }
  spinState.animationId = requestAnimationFrame(tick)
}

function stopMomentum() {
  if (spinState.animationId !== null) {
    cancelAnimationFrame(spinState.animationId)
    spinState.animationId = null
  }
}
</script>

<template>
  <header class="bg-dark px-4 py-3 flex items-center justify-between border-b border-dark-lighter">
    <div class="flex items-center gap-3">
      <div class="avatar-perspective">
        <div
          class="touch-none cursor-grab active:cursor-grabbing"
          :style="{ transform: `rotateY(${rotation}deg)` }"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
        >
          <UserAvatar
            :name="authStore.user?.name ?? ''"
            :cosmetics="userCosmetics"
            :rank-top3="true"
          />
        </div>
      </div>
      <div>
        <UserName
          v-if="authStore.user"
          :name="authStore.user.name"
          :cosmetics="userCosmetics"
        />
        <p class="text-sm text-gray-400">{{ titleText }}</p>
      </div>
    </div>
    <div class="flex items-center gap-2">
      <div 
        v-if="authStore.user?.coins" 
        @click="goToShop" 
        class="flex items-center gap-1 px-3 py-2 rounded-full bg-dark-light hover:bg-dark-lighter transition-colors cursor-pointer text-sm"
      >
        <span>🪙</span>
        <span class="text-yellow-300 font-bold">{{ authStore.user.coins }}</span>
      </div>
      <button
        @click="goToLeaderboard"
        class="bg-dark-light hover:bg-dark-lighter px-4 py-2 rounded-full flex items-center gap-2 transition-colors cursor-pointer"
      >
        <span class="text-primary font-bold">{{ authStore.user?.tokens ?? 0 }}</span>
        <span class="text-gray-400 text-sm">tokens</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.avatar-perspective {
  perspective: 200px;
}
</style>
