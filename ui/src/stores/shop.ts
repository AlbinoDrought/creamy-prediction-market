import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ShopItem } from '@/types/shop'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

export const useShopStore = defineStore('shop', () => {
  const allItems = ref<ShopItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const itemsByCategory = computed(() => {
    const map = new Map<string, ShopItem[]>()
    for (const item of allItems.value) {
      const list = map.get(item.category) || []
      list.push(item)
      map.set(item.category, list)
    }
    return map
  })

  function getItemById(id: string): ShopItem | undefined {
    return allItems.value.find(i => i.id === id)
  }

  async function fetchShopItems() {
    try {
      allItems.value = await api.getShopItems()
    } catch (e) {
      console.error('Failed to fetch shop items:', e)
    }
  }

  async function buyItem(itemId: string) {
    error.value = null
    try {
      await api.buyShopItem(itemId)
      // Refresh user data (coins + owned items)
      const authStore = useAuthStore()
      await authStore.swapUser()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Purchase failed'
      throw e
    }
  }

  async function equipItem(itemId: string) {
    error.value = null
    try {
      await api.equipItem(itemId)
      const authStore = useAuthStore()
      await authStore.swapUser()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Equip failed'
      throw e
    }
  }

  async function unequipCategory(category: string) {
    error.value = null
    try {
      await api.unequipCategory(category)
      const authStore = useAuthStore()
      await authStore.swapUser()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unequip failed'
      throw e
    }
  }

  return {
    allItems,
    loading,
    error,
    itemsByCategory,
    getItemById,
    fetchShopItems,
    buyItem,
    equipItem,
    unequipCategory,
  }
})
