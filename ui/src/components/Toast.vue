<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  message: string
  type?: 'success' | 'error' | 'info'
  duration?: number
}>()

const emit = defineEmits<{
  close: []
}>()

const visible = ref(true)

watch(
  () => props.message,
  () => {
    visible.value = true
    setTimeout(() => {
      visible.value = false
      emit('close')
    }, props.duration ?? 3000)
  },
  { immediate: true }
)
</script>

<template>
  <Transition name="toast">
    <div
      v-if="visible"
      class="fixed top-4 left-4 right-4 mx-auto max-w-md z-50 px-4 py-3 rounded-lg shadow-lg"
      :class="{
        'bg-success text-white': type === 'success',
        'bg-error text-white': type === 'error',
        'bg-dark-light text-white border border-dark-lighter': type === 'info' || !type,
      }"
    >
      <div class="flex items-center gap-3">
        <svg v-if="type === 'success'" class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <svg v-if="type === 'error'" class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
        <p class="flex-1">{{ message }}</p>
        <button @click="visible = false; emit('close')" class="text-white/80 hover:text-white">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}
</style>
