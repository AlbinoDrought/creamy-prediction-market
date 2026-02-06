<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { Achievement } from '@/types/achievements'

const props = defineProps<{
  achievement: Achievement
  duration?: number
}>()

const emit = defineEmits<{
  close: []
}>()

const visible = ref(false)
const iconRevealed = ref(false)

onMounted(() => {
  // Stagger the entrance animations
  requestAnimationFrame(() => {
    visible.value = true
    setTimeout(() => {
      iconRevealed.value = true
    }, 300)
  })

  setTimeout(() => {
    visible.value = false
    setTimeout(() => emit('close'), 500)
  }, props.duration ?? 5000)
})

function dismiss() {
  visible.value = false
  setTimeout(() => emit('close'), 500)
}
</script>

<template>
  <Teleport to="body">
    <div
      class="achievement-overlay"
      :class="{ 'is-visible': visible }"
      @click="dismiss"
    >
      <div class="achievement-card" :class="{ 'is-visible': visible }" @click.stop="dismiss">
        <!-- Shimmer sweep -->
        <div class="achievement-shimmer" />

        <!-- Glowing border -->
        <div class="achievement-border" />

        <!-- Content -->
        <div class="achievement-content">
          <!-- Icon with glow ring -->
          <div class="achievement-icon-wrapper" :class="{ revealed: iconRevealed }">
            <div class="achievement-glow-ring" />
            <div class="achievement-glow-ring ring-2" />
            <span class="achievement-icon">{{ achievement.icon }}</span>
          </div>

          <!-- Header -->
          <p class="achievement-header">ACHIEVEMENT UNLOCKED</p>

          <!-- Name -->
          <p class="achievement-name">{{ achievement.name }}</p>

          <!-- Description -->
          <p class="achievement-description">{{ achievement.description }}</p>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.achievement-overlay {
  position: fixed;
  inset: 0;
  z-index: 9998;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0);
  transition: background 0.4s ease;
  pointer-events: none;
}

.achievement-overlay.is-visible {
  background: rgba(0, 0, 0, 0.6);
  pointer-events: auto;
}

.achievement-card {
  position: relative;
  width: 320px;
  padding: 2px;
  border-radius: 20px;
  opacity: 0;
  transform: scale(0.5) translateY(30px);
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  cursor: pointer;
}

.achievement-card.is-visible {
  opacity: 1;
  transform: scale(1) translateY(0);
}

/* Animated gradient border */
.achievement-border {
  position: absolute;
  inset: 0;
  border-radius: 20px;
  background: conic-gradient(
    from var(--border-angle, 0deg),
    #facc15,
    #7c3aed,
    #ec4899,
    #facc15
  );
  animation: rotate-border 3s linear infinite;
}

@property --border-angle {
  syntax: "<angle>";
  initial-value: 0deg;
  inherits: false;
}

@keyframes rotate-border {
  to {
    --border-angle: 360deg;
  }
}

/* Card content sits on top of border */
.achievement-content {
  position: relative;
  z-index: 1;
  background: linear-gradient(145deg, #1f2937, #111827);
  border-radius: 18px;
  padding: 32px 24px 28px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  overflow: hidden;
}

/* Shimmer effect */
.achievement-shimmer {
  position: absolute;
  inset: 0;
  z-index: 2;
  border-radius: 20px;
  overflow: hidden;
  pointer-events: none;
}

.achievement-shimmer::after {
  content: '';
  position: absolute;
  top: -50%;
  left: -100%;
  width: 60%;
  height: 200%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.08),
    rgba(255, 255, 255, 0.15),
    rgba(255, 255, 255, 0.08),
    transparent
  );
  transform: skewX(-20deg);
  animation: shimmer 3s ease-in-out 0.5s infinite;
}

@keyframes shimmer {
  0% {
    left: -100%;
  }
  40%, 100% {
    left: 200%;
  }
}

/* Icon */
.achievement-icon-wrapper {
  position: relative;
  width: 88px;
  height: 88px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
  opacity: 0;
  transform: scale(0);
  transition: all 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.achievement-icon-wrapper.revealed {
  opacity: 1;
  transform: scale(1);
}

.achievement-icon {
  font-size: 48px;
  position: relative;
  z-index: 1;
  filter: drop-shadow(0 0 12px rgba(250, 204, 21, 0.5));
}

.achievement-glow-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 2px solid rgba(250, 204, 21, 0.4);
  animation: pulse-ring 2s ease-in-out infinite;
}

.achievement-glow-ring.ring-2 {
  inset: -8px;
  border-color: rgba(124, 58, 237, 0.3);
  animation-delay: 0.5s;
}

@keyframes pulse-ring {
  0%, 100% {
    opacity: 0.4;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.1);
  }
}

/* Text */
.achievement-header {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 3px;
  color: #facc15;
  margin-bottom: 8px;
  text-shadow: 0 0 20px rgba(250, 204, 21, 0.4);
}

.achievement-name {
  font-size: 22px;
  font-weight: 800;
  color: #fff;
  margin-bottom: 8px;
  line-height: 1.2;
}

.achievement-description {
  font-size: 14px;
  color: #9ca3af;
  line-height: 1.4;
}
</style>
