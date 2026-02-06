import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Achievement, UserAchievement } from '@/types/achievements'
import { api } from '@/api/client'
import { useConfetti } from '@/composables/useConfetti'

export const useAchievementsStore = defineStore('achievements', () => {
  const allAchievements = ref<Achievement[]>([])
  const myAchievements = ref<UserAchievement[]>([])
  const newlyEarnedAchievement = ref<Achievement | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const { fire: fireConfetti } = useConfetti()

  async function fetchAchievements() {
    try {
      allAchievements.value = await api.getAchievements()
    } catch (e) {
      console.error('Failed to fetch achievements:', e)
    }
  }

  async function fetchMyAchievements() {
    loading.value = true
    error.value = null
    try {
      myAchievements.value = await api.getMyAchievements()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load achievements'
    } finally {
      loading.value = false
    }
  }

  function getAchievementById(id: string): Achievement | undefined {
    return allAchievements.value.find(a => a.id === id)
  }

  function onAchievementEarned(achievementId: string) {
    const achievement = getAchievementById(achievementId)
    if (achievement) {
      newlyEarnedAchievement.value = achievement
      fireConfetti(100)
      // Also refresh the list
      fetchMyAchievements()
    }
  }

  function clearNewlyEarned() {
    newlyEarnedAchievement.value = null
  }

  return {
    allAchievements,
    myAchievements,
    newlyEarnedAchievement,
    loading,
    error,
    fetchAchievements,
    fetchMyAchievements,
    getAchievementById,
    onAchievementEarned,
    clearNewlyEarned,
  }
})
