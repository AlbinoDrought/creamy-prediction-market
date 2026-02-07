<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import { onSSEEvent } from '@/composables/useSSE'
import type { UserCosmetics } from '@/types/users'
import AppHeader from '@/components/AppHeader.vue'
import BottomNav from '@/components/BottomNav.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import UserName from '@/components/UserName.vue'

const authStore = useAuthStore()

const canvas = ref<HTMLCanvasElement | null>(null)
const gameState = ref<'idle' | 'playing' | 'over'>('idle')
const score = ref(0)
const highScore = ref(parseInt(localStorage.getItem('minigame_high') || '0', 10))
const coinsEarned = ref(0)
const claiming = ref(false)

interface MinigameEntry {
  name: string
  high_score: number
  cosmetics: UserCosmetics
}
const leaderboard = ref<MinigameEntry[]>([])

async function fetchLeaderboard() {
  try {
    leaderboard.value = await api.getMinigameLeaderboard()
  } catch { /* ignore */ }
}

// Game constants
const GROUND_Y = 0.75 // ground at 75% of canvas height
const DINO_SIZE = 0.06 // relative to canvas width
const GRAVITY = 0.0015
const JUMP_VELOCITY = -0.028
const OBSTACLE_WIDTH = 0.03
const OBSTACLE_GAP_MIN = 0.4
const OBSTACLE_GAP_MAX = 0.7

// Game state (not reactive for performance)
let animId: number | null = null
let dino = { x: 0, y: 0, vy: 0, size: 0, grounded: true }
let obstacles: { x: number; height: number; width: number }[] = []
let baseSpeed = 0
let speedMultiplier = 0
let nextObstacleIn = 0
let groundY = 0
let frameScore = 0
let lastTime = 0

// Logical (CSS) size of the canvas — used for all game math and drawing
let canvasW = 0
let canvasH = 0

function getCanvasSize() {
  return { w: canvasW, h: canvasH }
}

function resetGame() {
  const { w, h } = getCanvasSize()
  groundY = h * GROUND_Y
  dino = {
    x: w * 0.1,
    y: groundY,
    vy: 0,
    size: w * DINO_SIZE,
    grounded: true,
  }
  obstacles = []
  baseSpeed = w * 0.003
  speedMultiplier = 1
  nextObstacleIn = w * 0.15
  frameScore = 0
  score.value = 0
  lastTime = 0
}

function jump() {
  if (gameState.value === 'idle' || gameState.value === 'over') {
    gameState.value = 'playing'
    resetGame()
    startLoop()
    return
  }
  if (gameState.value === 'playing' && dino.grounded) {
    dino.vy = JUMP_VELOCITY * getCanvasSize().h
    dino.grounded = false
  }
}

function startLoop() {
  lastTime = performance.now()
  function loop(time: number) {
    const dt = Math.min((time - lastTime) / 16.67, 3) // normalize to ~60fps, cap at 3x
    lastTime = time
    update(dt)
    draw()
    if (gameState.value === 'playing') {
      animId = requestAnimationFrame(loop)
    }
  }
  animId = requestAnimationFrame(loop)
}

function update(dt: number) {
  const { w, h } = getCanvasSize()
  const speed = baseSpeed * speedMultiplier * dt

  // Dino physics
  if (!dino.grounded) {
    dino.vy += GRAVITY * h * dt
    dino.y += dino.vy * dt
    if (dino.y >= groundY) {
      dino.y = groundY
      dino.vy = 0
      dino.grounded = true
    }
  }

  // Move obstacles
  for (const obs of obstacles) {
    obs.x -= speed
  }

  // Remove off-screen obstacles
  obstacles = obstacles.filter(o => o.x + o.width > 0)

  // Spawn obstacles
  nextObstacleIn -= speed
  if (nextObstacleIn <= 0) {
    const obsW = w * OBSTACLE_WIDTH
    const obsH = h * (0.06 + Math.random() * 0.1)
    obstacles.push({ x: w + obsW, height: obsH, width: obsW })
    nextObstacleIn = w * (OBSTACLE_GAP_MIN + Math.random() * (OBSTACLE_GAP_MAX - OBSTACLE_GAP_MIN))
  }

  // Collision detection
  const dinoLeft = dino.x - dino.size * 0.4
  const dinoRight = dino.x + dino.size * 0.4
  const dinoTop = dino.y - dino.size
  const dinoBottom = dino.y
  for (const obs of obstacles) {
    const obsLeft = obs.x
    const obsRight = obs.x + obs.width
    const obsTop = groundY - obs.height
    const obsBottom = groundY
    if (dinoRight > obsLeft && dinoLeft < obsRight && dinoBottom > obsTop && dinoTop < obsBottom) {
      gameOver()
      return
    }
  }

  // Score & difficulty
  frameScore += speed
  score.value = Math.floor(frameScore / 5)
  speedMultiplier = 1 + score.value / 500
}

async function gameOver() {
  gameState.value = 'over'
  if (animId !== null) {
    cancelAnimationFrame(animId)
    animId = null
  }

  if (score.value > highScore.value) {
    highScore.value = score.value
    localStorage.setItem('minigame_high', String(score.value))
  }

  // Claim coins
  claiming.value = true
  try {
    const result = await api.claimMinigameCoins(score.value)
    coinsEarned.value = result.coins_earned
    await authStore.swapUser()
    fetchLeaderboard()
  } catch {
    coinsEarned.value = 0
  } finally {
    claiming.value = false
  }
}

function draw() {
  const el = canvas.value
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return
  const { w, h } = getCanvasSize()

  // Clear
  ctx.fillStyle = '#111827'
  ctx.fillRect(0, 0, w, h)

  // Ground
  ctx.strokeStyle = '#4B5563'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(0, groundY + 1)
  ctx.lineTo(w, groundY + 1)
  ctx.stroke()

  // Ground texture (small dashes)
  ctx.strokeStyle = '#374151'
  ctx.lineWidth = 1
  for (let x = (frameScore * 0.5) % 20; x < w; x += 20) {
    ctx.beginPath()
    ctx.moveTo(w - x, groundY + 6)
    ctx.lineTo(w - x + 8, groundY + 6)
    ctx.stroke()
  }

  // Dino (simple T-rex shape)
  const s = dino.size
  ctx.fillStyle = '#22C55E'
  // Body
  ctx.fillRect(dino.x - s * 0.3, dino.y - s * 0.85, s * 0.6, s * 0.7)
  // Head
  ctx.fillRect(dino.x, dino.y - s, s * 0.5, s * 0.4)
  // Eye
  ctx.fillStyle = '#111827'
  ctx.fillRect(dino.x + s * 0.3, dino.y - s * 0.9, s * 0.1, s * 0.1)
  // Legs
  ctx.fillStyle = '#22C55E'
  if (dino.grounded) {
    // Alternate legs based on score
    const legOffset = Math.floor(frameScore / 10) % 2 === 0
    ctx.fillRect(dino.x - s * 0.15, dino.y - s * 0.15, s * 0.12, s * 0.15)
    ctx.fillRect(dino.x + s * 0.1, dino.y - s * 0.15, s * 0.12, legOffset ? s * 0.15 : s * 0.1)
  } else {
    // Tucked legs when jumping
    ctx.fillRect(dino.x - s * 0.1, dino.y - s * 0.2, s * 0.1, s * 0.1)
    ctx.fillRect(dino.x + s * 0.05, dino.y - s * 0.2, s * 0.1, s * 0.1)
  }
  // Tail
  ctx.fillRect(dino.x - s * 0.55, dino.y - s * 0.7, s * 0.3, s * 0.2)

  // Obstacles (cacti)
  ctx.fillStyle = '#EF4444'
  for (const obs of obstacles) {
    // Main body
    ctx.fillRect(obs.x, groundY - obs.height, obs.width, obs.height)
    // Small arms
    const armW = obs.width * 0.4
    const armH = obs.height * 0.25
    ctx.fillRect(obs.x - armW, groundY - obs.height * 0.7, armW, armH)
    ctx.fillRect(obs.x + obs.width, groundY - obs.height * 0.5, armW, armH)
  }

  // Score display
  ctx.fillStyle = '#9CA3AF'
  ctx.font = `bold ${Math.floor(h * 0.05)}px monospace`
  ctx.textAlign = 'right'
  ctx.fillText(String(score.value).padStart(5, '0'), w - 16, h * 0.08)

  // High score
  if (highScore.value > 0) {
    ctx.fillStyle = '#6B7280'
    ctx.font = `${Math.floor(h * 0.035)}px monospace`
    ctx.fillText(`HI ${String(highScore.value).padStart(5, '0')}`, w - 16, h * 0.13)
  }
}

function drawIdleScreen() {
  const el = canvas.value
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return
  const { w, h } = getCanvasSize()

  ctx.fillStyle = '#111827'
  ctx.fillRect(0, 0, w, h)

  // Ground
  groundY = h * GROUND_Y
  ctx.strokeStyle = '#4B5563'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(0, groundY + 1)
  ctx.lineTo(w, groundY + 1)
  ctx.stroke()

  // Static dino
  const s = w * DINO_SIZE
  const dx = w * 0.1
  ctx.fillStyle = '#22C55E'
  ctx.fillRect(dx - s * 0.3, groundY - s * 0.85, s * 0.6, s * 0.7)
  ctx.fillRect(dx, groundY - s, s * 0.5, s * 0.4)
  ctx.fillStyle = '#111827'
  ctx.fillRect(dx + s * 0.3, groundY - s * 0.9, s * 0.1, s * 0.1)
  ctx.fillStyle = '#22C55E'
  ctx.fillRect(dx - s * 0.15, groundY - s * 0.15, s * 0.12, s * 0.15)
  ctx.fillRect(dx + s * 0.1, groundY - s * 0.15, s * 0.12, s * 0.15)
  ctx.fillRect(dx - s * 0.55, groundY - s * 0.7, s * 0.3, s * 0.2)

  // Prompt
  ctx.fillStyle = '#9CA3AF'
  ctx.font = `bold ${Math.floor(h * 0.06)}px monospace`
  ctx.textAlign = 'center'
  ctx.fillText('TAP TO START', w / 2, h * 0.45)

  ctx.fillStyle = '#6B7280'
  ctx.font = `${Math.floor(h * 0.035)}px monospace`
  ctx.fillText('Earn coins for the shop!', w / 2, h * 0.52)
  ctx.fillText('1 coin per 100 pts (max 5)', w / 2, h * 0.57)
}

function resizeCanvas() {
  const el = canvas.value
  if (!el) return
  const parent = el.parentElement
  if (!parent) return
  const dpr = window.devicePixelRatio || 1
  const rect = parent.getBoundingClientRect()
  canvasW = rect.width
  canvasH = rect.height
  el.width = rect.width * dpr
  el.height = rect.height * dpr
  el.style.width = `${rect.width}px`
  el.style.height = `${rect.height}px`
  const ctx = el.getContext('2d')
  if (ctx) ctx.scale(dpr, dpr)

  if (gameState.value === 'idle') {
    drawIdleScreen()
  } else if (gameState.value === 'over') {
    draw()
  }
}

function onKeyDown(e: KeyboardEvent) {
  if (e.code === 'Space' || e.code === 'ArrowUp') {
    e.preventDefault()
    jump()
  }
}

let unsubSSE: (() => void) | null = null

onMounted(() => {
  resizeCanvas()
  drawIdleScreen()
  window.addEventListener('resize', resizeCanvas)
  window.addEventListener('keydown', onKeyDown)
  fetchLeaderboard()
  unsubSSE = onSSEEvent((event) => {
    if (event.type === 'minigame_leaderboard') {
      fetchLeaderboard()
    }
  })
})

onUnmounted(() => {
  if (animId !== null) {
    cancelAnimationFrame(animId)
    animId = null
  }
  window.removeEventListener('resize', resizeCanvas)
  window.removeEventListener('keydown', onKeyDown)
  unsubSSE?.()
})
</script>

<template>
  <div class="min-h-screen bg-bg pb-20">
    <AppHeader />

    <main class="p-4 flex flex-col gap-4">
      <div class="relative w-full aspect-[2/1] bg-dark rounded-xl overflow-hidden">
        <canvas
          ref="canvas"
          class="w-full h-full block"
          @click="jump"
          @touchstart.prevent="jump"
        />

        <!-- Game Over Overlay -->
        <div
          v-if="gameState === 'over'"
          class="absolute inset-0 flex flex-col items-center justify-center bg-black/50"
          @click="jump"
          @touchstart.prevent="jump"
        >
          <p class="text-2xl font-bold text-white mb-1">GAME OVER</p>
          <p class="text-lg text-gray-300 mb-1">Score: {{ score }}</p>
          <p v-if="claiming" class="text-sm text-gray-400">Claiming coins...</p>
          <p v-else-if="coinsEarned > 0" class="text-yellow-300 font-bold mb-3">+{{ coinsEarned }} coins earned!</p>
          <p v-else class="text-gray-400 text-sm mb-3">Score 100+ to earn coins</p>
          <p class="text-sm text-gray-400">Tap to play again</p>
        </div>
      </div>

      <div class="flex justify-between items-center text-sm text-gray-400">
        <span>High Score: {{ highScore }}</span>
        <span>Coins: 🪙 {{ authStore.user?.coins ?? 0 }}</span>
      </div>

      <!-- Leaderboard -->
      <div v-if="leaderboard.length" class="mt-2">
        <h2 class="text-lg font-bold text-white mb-3">High Scores</h2>
        <div class="flex flex-col gap-2">
          <div
            v-for="(entry, idx) in leaderboard"
            :key="entry.name"
            class="flex items-center gap-3 bg-dark-light rounded-xl px-4 py-3"
            :class="{ 'ring-1 ring-primary': entry.name === authStore.user?.name }"
          >
            <span class="text-sm font-bold w-8 shrink-0" :class="idx === 0 ? 'text-yellow-400' : idx === 1 ? 'text-gray-300' : idx === 2 ? 'text-amber-600' : 'text-gray-500'">
              {{ idx + 1 }}.
            </span>
            <UserAvatar :name="entry.name" :cosmetics="entry.cosmetics" size="sm" />
            <UserName :name="entry.name" :cosmetics="entry.cosmetics" class="flex-1 truncate" />
            <span class="text-sm font-mono font-bold text-white">{{ entry.high_score }}</span>
          </div>
        </div>
      </div>
    </main>

    <BottomNav />
  </div>
</template>
