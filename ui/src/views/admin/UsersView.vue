<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { User } from '@/types/users'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import Toast from '@/components/Toast.vue'

const users = ref<User[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const showToast = ref(false)
const actionLoading = ref<string | null>(null)

const giftModal = ref<{ userId: string; userName: string } | null>(null)
const giftAmount = ref(100)

const resetModal = ref<{ userId: string; userName: string } | null>(null)
const newPin = ref('')

onMounted(async () => {
  await fetchUsers()
})

async function fetchUsers() {
  loading.value = true
  try {
    users.value = await api.getUsers()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load users'
  } finally {
    loading.value = false
  }
}

function openGiftModal(user: User) {
  giftModal.value = { userId: user.id, userName: user.name }
  giftAmount.value = 100
}

async function giftTokens() {
  if (!giftModal.value || giftAmount.value <= 0) return

  actionLoading.value = giftModal.value.userId
  try {
    await api.giftTokens(giftModal.value.userId, giftAmount.value)
    await fetchUsers()
    toastType.value = 'success'
    toastMessage.value = `Gifted ${giftAmount.value} tokens to ${giftModal.value.userName}`
    showToast.value = true
    giftModal.value = null
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to gift tokens'
    showToast.value = true
  } finally {
    actionLoading.value = null
  }
}

function openResetModal(user: User) {
  resetModal.value = { userId: user.id, userName: user.name }
  newPin.value = ''
}

async function resetPin() {
  if (!resetModal.value || !newPin.value.trim()) return

  actionLoading.value = resetModal.value.userId
  try {
    await api.resetUserPin(resetModal.value.userId, newPin.value)
    toastType.value = 'success'
    toastMessage.value = `PIN reset for ${resetModal.value.userName}`
    showToast.value = true
    resetModal.value = null
  } catch (e) {
    toastType.value = 'error'
    toastMessage.value = e instanceof Error ? e.message : 'Failed to reset PIN'
    showToast.value = true
  } finally {
    actionLoading.value = null
  }
}
</script>

<template>
  <div>
    <h2 class="text-xl font-bold text-white mb-6">Users</h2>

    <div v-if="loading" class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>

    <div v-else-if="error" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error">
      {{ error }}
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="user in users"
        :key="user.id"
        class="bg-dark-light rounded-xl p-4 flex items-center justify-between"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 rounded-full flex items-center justify-center font-bold"
            :class="user.admin
              ? 'bg-secondary text-white'
              : 'bg-gradient-to-br from-primary to-secondary text-dark'"
          >
            {{ user.name.charAt(0).toUpperCase() }}
          </div>
          <div>
            <p class="font-medium text-white">
              {{ user.name }}
              <span v-if="user.admin" class="text-xs text-secondary ml-1">(admin)</span>
            </p>
            <p class="text-sm text-gray-400">{{ user.tokens }} tokens</p>
          </div>
        </div>

        <div class="flex gap-2">
          <button
            @click="openGiftModal(user)"
            class="bg-dark hover:bg-dark-lighter text-primary px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            Gift
          </button>
          <button
            @click="openResetModal(user)"
            class="bg-dark hover:bg-dark-lighter text-gray-400 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            Reset PIN
          </button>
        </div>
      </div>

      <div v-if="users.length === 0" class="text-center py-12">
        <p class="text-gray-400">No users yet</p>
      </div>
    </div>

    <!-- Gift Modal -->
    <Teleport to="body">
      <div
        v-if="giftModal"
        class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50"
        @click.self="giftModal = null"
      >
        <div class="bg-dark rounded-xl p-6 w-full max-w-sm">
          <h3 class="text-lg font-bold text-white mb-4">
            Gift Tokens to {{ giftModal.userName }}
          </h3>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-300 mb-2">Amount</label>
            <input
              v-model.number="giftAmount"
              type="number"
              min="1"
              class="w-full bg-dark-light border border-dark-lighter rounded-lg px-4 py-3 text-white focus:outline-none focus:border-secondary"
            />
          </div>
          <div class="flex gap-2">
            <button
              @click="giftTokens"
              :disabled="actionLoading !== null || giftAmount <= 0"
              class="flex-1 bg-primary hover:bg-primary-dark disabled:opacity-50 text-dark font-bold py-3 rounded-lg transition-colors"
            >
              {{ actionLoading ? 'Gifting...' : 'Gift Tokens' }}
            </button>
            <button
              @click="giftModal = null"
              class="px-4 text-gray-400 hover:text-white"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Reset PIN Modal -->
    <Teleport to="body">
      <div
        v-if="resetModal"
        class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50"
        @click.self="resetModal = null"
      >
        <div class="bg-dark rounded-xl p-6 w-full max-w-sm">
          <h3 class="text-lg font-bold text-white mb-4">
            Reset PIN for {{ resetModal.userName }}
          </h3>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-300 mb-2">New PIN</label>
            <input
              v-model="newPin"
              type="password"
              inputmode="numeric"
              pattern="[0-9]*"
              class="w-full bg-dark-light border border-dark-lighter rounded-lg px-4 py-3 text-white focus:outline-none focus:border-secondary"
              placeholder="Enter new PIN"
            />
          </div>
          <div class="flex gap-2">
            <button
              @click="resetPin"
              :disabled="actionLoading !== null || !newPin.trim()"
              class="flex-1 bg-secondary hover:bg-secondary-dark disabled:opacity-50 text-white font-bold py-3 rounded-lg transition-colors"
            >
              {{ actionLoading ? 'Resetting...' : 'Reset PIN' }}
            </button>
            <button
              @click="resetModal = null"
              class="px-4 text-gray-400 hover:text-white"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <Toast
      v-if="showToast"
      :message="toastMessage"
      :type="toastType"
      @close="showToast = false"
    />
  </div>
</template>
