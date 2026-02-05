import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/users'
import { api } from '@/api/client'

const STORAGE_KEY = 'auth_token'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem(STORAGE_KEY))
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.admin ?? false)

  // Initialize API client with stored token
  if (token.value) {
    api.setToken(token.value)
  }

  function setAuth(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem(STORAGE_KEY, newToken)
    api.setToken(newToken)
  }

  function clearAuth() {
    token.value = null
    user.value = null
    localStorage.removeItem(STORAGE_KEY)
    api.setToken(null)
  }

  async function register(name: string, pin: string) {
    loading.value = true
    error.value = null
    try {
      const result = await api.register(name, pin)
      setAuth(result.token, result.user)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Registration failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function login(name: string, pin: string) {
    loading.value = true
    error.value = null
    try {
      const result = await api.login(name, pin)
      setAuth(result.token, result.user)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Login failed'
      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchUser() {
    if (!token.value) return false
    loading.value = true
    try {
      user.value = await api.getMe()
      return true
    } catch {
      clearAuth()
      return false
    } finally {
      loading.value = false
    }
  }

  function logout() {
    clearAuth()
  }

  function updateTokens(newTokens: number) {
    if (user.value) {
      user.value = { ...user.value, tokens: newTokens }
    }
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    isAdmin,
    register,
    login,
    logout,
    fetchUser,
    updateTokens,
  }
})
