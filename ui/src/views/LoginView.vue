<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

const authStore = useAuthStore()
const router = useRouter()

const isLogin = ref(true)
const name = ref('')
const pin = ref('')

async function handleSubmit() {
  if (!name.value.trim() || !pin.value.trim()) return

  const success = isLogin.value
    ? await authStore.login(name.value.trim(), pin.value)
    : await authStore.register(name.value.trim(), pin.value)

  if (success) {
    router.push({ name: 'home' })
  }
}
</script>

<template>
  <div class="min-h-screen bg-bg flex flex-col items-center justify-center p-6">
    <div class="w-full max-w-sm">
      <!-- Logo/Title -->
      <div class="text-center mb-8">
        <div class="w-20 h-20 mx-auto mb-4 rounded-full bg-gradient-to-br from-primary to-secondary flex items-center justify-center">
          <svg class="w-10 h-10 text-dark" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-white">Superbowl Predictions</h1>
        <p class="text-gray-400 mt-2">Bet on the game with your friends</p>
      </div>

      <!-- Toggle -->
      <div class="flex bg-dark rounded-lg p-1 mb-6">
        <button
          @click="isLogin = true"
          class="flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors"
          :class="isLogin ? 'bg-primary text-dark' : 'text-gray-400 hover:text-white'"
        >
          Sign In
        </button>
        <button
          @click="isLogin = false"
          class="flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors"
          :class="!isLogin ? 'bg-primary text-dark' : 'text-gray-400 hover:text-white'"
        >
          Register
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label for="name" class="block text-sm font-medium text-gray-300 mb-2">
            Username
          </label>
          <input
            id="name"
            v-model="name"
            type="text"
            autocomplete="username"
            required
            class="w-full bg-dark border border-dark-lighter rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-primary transition-colors"
            placeholder="Enter your name"
          />
        </div>

        <div>
          <label for="pin" class="block text-sm font-medium text-gray-300 mb-2">
            PIN
          </label>
          <input
            id="pin"
            v-model="pin"
            type="password"
            inputmode="numeric"
            pattern="[0-9]*"
            autocomplete="current-password"
            required
            class="w-full bg-dark border border-dark-lighter rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-primary transition-colors"
            placeholder="4-digit PIN"
          />
        </div>

        <div v-if="authStore.error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error text-sm">
          {{ authStore.error }}
        </div>

        <button
          type="submit"
          :disabled="authStore.loading || !name.trim() || !pin.trim()"
          class="w-full bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-dark font-bold py-3 px-4 rounded-lg transition-colors flex items-center justify-center gap-2"
        >
          <LoadingSpinner v-if="authStore.loading" size="sm" />
          <span v-else>{{ isLogin ? 'Sign In' : 'Create Account' }}</span>
        </button>
      </form>
    </div>
  </div>
</template>
