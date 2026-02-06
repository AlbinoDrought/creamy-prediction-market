<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useShopStore } from '@/stores/shop'
import { ShopItemCategory } from '@/types/shop'
import type { ShopItem } from '@/types/shop'
import AppHeader from '@/components/AppHeader.vue'
import BottomNav from '@/components/BottomNav.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import UserName from '@/components/UserName.vue'

const authStore = useAuthStore()
const shopStore = useShopStore()

const buyingId = ref<string | null>(null)
const buyError = ref<string | null>(null)

// Confirmation modal state
const confirmItem = ref<ShopItem | null>(null)
const confirmMode = ref<'buy' | 'equip'>('buy')

onMounted(() => {
  shopStore.fetchShopItems()
})

const categoryOrder = [
  { key: ShopItemCategory.Hat, label: 'Hats' },
  { key: ShopItemCategory.AvatarEffect, label: 'Avatar Effects' },
  { key: ShopItemCategory.AvatarEmoji, label: 'Avatar Emojis' },
  { key: ShopItemCategory.AvatarColor, label: 'Avatar Colors' },
  { key: ShopItemCategory.NameEffect, label: 'Name Effects' },
  { key: ShopItemCategory.NameEmoji, label: 'Name Emojis' },
  { key: ShopItemCategory.NameFont, label: 'Name Fonts' },
  { key: ShopItemCategory.NameBold, label: 'Name Style' },
  { key: ShopItemCategory.Title, label: 'Titles' },
  { key: ShopItemCategory.GlobalAction, label: 'Global Actions' },
]

const userCoins = computed(() => authStore.user?.coins ?? 0)
const ownedItems = computed(() => new Set(authStore.user?.owned_items ?? []))
const userCosmetics = computed(() => authStore.user?.cosmetics || {
  avatar_color: '',
  avatar_emoji: '',
  name_emoji: '',
  avatar_effect: '',
  name_effect: '',
  name_bold: false,
  name_font: '',
  title: '',
  hat: '',
})

function isOwned(item: ShopItem): boolean {
  return ownedItems.value.has(item.id)
}

function isEquipped(item: ShopItem): boolean {
  const c = userCosmetics.value
  if (!c) return false
  switch (item.category) {
    case ShopItemCategory.AvatarColor: return c.avatar_color === item.value
    case ShopItemCategory.AvatarEmoji: return c.avatar_emoji === item.value
    case ShopItemCategory.NameEmoji: return c.name_emoji === item.value
    case ShopItemCategory.AvatarEffect: return c.avatar_effect === item.value
    case ShopItemCategory.NameEffect: return c.name_effect === item.value
    case ShopItemCategory.NameBold: return c.name_bold === true
    case ShopItemCategory.NameFont: return c.name_font === item.value
    case ShopItemCategory.Title: return c.title === item.value
    case ShopItemCategory.Hat: return c.hat === item.value
    default: return false
  }
}

function canAfford(item: ShopItem): boolean {
  return userCoins.value >= item.price
}

function sortedItems(category: ShopItemCategory): ShopItem[] {
  const items = shopStore.itemsByCategory.get(category) ?? []
  return [...items].sort((a, b) => a.name.localeCompare(b.name))
}

// Preview cosmetics: what the user would look like with this item
function previewCosmetics(item: ShopItem) {
  const base = { ...(userCosmetics.value || {}) }
  switch (item.category) {
    case ShopItemCategory.AvatarColor: base.avatar_color = item.value; break
    case ShopItemCategory.AvatarEmoji: base.avatar_emoji = item.value; break
    case ShopItemCategory.NameEmoji: base.name_emoji = item.value; break
    case ShopItemCategory.AvatarEffect: base.avatar_effect = item.value; break
    case ShopItemCategory.NameEffect: base.name_effect = item.value; break
    case ShopItemCategory.NameBold: base.name_bold = true; break
    case ShopItemCategory.NameFont: base.name_font = item.value; break
    case ShopItemCategory.Title: base.title = item.value; break
    case ShopItemCategory.Hat: base.hat = item.value; break
  }
  return base
}

function promptBuy(item: ShopItem) {
  buyError.value = null
  confirmItem.value = item
  confirmMode.value = 'buy'
}

function promptEquip(item: ShopItem) {
  buyError.value = null
  confirmItem.value = item
  confirmMode.value = 'equip'
}

function cancelBuy() {
  confirmItem.value = null
}

async function confirmBuy() {
  const item = confirmItem.value
  if (!item) return
  confirmItem.value = null
  buyError.value = null
  buyingId.value = item.id
  try {
    await shopStore.buyItem(item.id)
    // Auto-equip non-consumable items after purchase
    if (!item.consumable) {
      await shopStore.equipItem(item.id)
    }
  } catch (e) {
    buyError.value = e instanceof Error ? e.message : 'Purchase failed'
  } finally {
    buyingId.value = null
  }
}

async function toggleEquip(item: ShopItem) {
  buyError.value = null
  buyingId.value = item.id
  try {
    if (isEquipped(item)) {
      await shopStore.unequipCategory(item.category)
    } else {
      await shopStore.equipItem(item.id)
    }
  } catch (e) {
    buyError.value = e instanceof Error ? e.message : 'Action failed'
  } finally {
    buyingId.value = null
  }
}

async function confirmEquip() {
  const item = confirmItem.value
  if (!item) return
  confirmItem.value = null
  await toggleEquip(item)
}

function onItemClick(item: ShopItem) {
  if (isButtonDisabled(item)) return
  if (item.consumable) {
    promptBuy(item)
  } else if (!isOwned(item)) {
    promptBuy(item)
  } else {
    promptEquip(item)
  }
}

function buttonLabel(item: ShopItem) {
  if (item.consumable) return `Use · ${item.price} 🪙`
  if (!isOwned(item)) return `${item.price} 🪙`
  if (isEquipped(item)) return 'Equipped'
  return 'Equip'
}

function isButtonDisabled(item: ShopItem) {
  if (buyingId.value === item.id) return true
  if (item.consumable) return !canAfford(item)
  if (!isOwned(item)) return !canAfford(item)
  return false
}

const confirmPreviewCosmetics = computed(() => {
  if (!confirmItem.value) return userCosmetics.value
  return previewCosmetics(confirmItem.value)
})
</script>

<template>
  <div class="min-h-screen bg-bg pb-20">
    <AppHeader />

    <main class="p-4">
      <!-- Error -->
      <div v-if="buyError" class="bg-error/20 border border-error rounded-lg px-4 py-3 text-error text-sm mb-4">
        {{ buyError }}
      </div>

      <!-- Category sections -->
      <div v-for="cat in categoryOrder" :key="cat.key" class="mb-6">
        <template v-if="shopStore.itemsByCategory.get(cat.key)?.length">
          <h2 class="text-lg font-bold text-white mb-3">{{ cat.label }}</h2>
          <div class="grid grid-cols-2 gap-3">
            <div
              v-for="item in sortedItems(cat.key)"
              :key="item.id"
              class="bg-dark-light rounded-xl p-3 flex flex-col items-center gap-2 relative"
              :class="{
                'ring-2 ring-primary': isEquipped(item),
                'opacity-60': !isOwned(item) && !canAfford(item),
              }"
              @click.prevent="onItemClick(item)"
            >
              <span class="text-2xl">{{ item.icon }}</span>
              <span class="text-sm font-medium text-white text-center">{{ item.name }}</span>

              <button
                class="w-full text-xs font-bold py-1.5 px-3 rounded-lg transition-colors min-h-0"
                :class="
                  isEquipped(item)
                    ? 'bg-primary text-dark'
                    : isOwned(item)
                    ? 'bg-dark-lighter text-white hover:bg-primary hover:text-dark'
                    : canAfford(item)
                    ? 'bg-yellow-500/20 text-yellow-300 hover:bg-yellow-500/30'
                    : 'bg-dark-lighter text-gray-500 cursor-not-allowed'
                "
                :disabled="isButtonDisabled(item)"
                @click.prevent="onItemClick(item)"
              >
                <template v-if="buyingId === item.id">...</template>
                <template v-else>{{ buttonLabel(item) }}</template>
              </button>
            </div>
          </div>
        </template>
      </div>
    </main>

    <BottomNav />

    <!-- Confirmation Modal -->
    <Teleport to="body">
      <div
        v-if="confirmItem"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @click.self="cancelBuy"
      >
        <div class="absolute inset-0 bg-black/60" @click="cancelBuy" />
        <div class="relative bg-dark-light rounded-2xl p-6 w-full max-w-sm shadow-xl">
          <!-- Preview -->
          <div class="flex flex-col items-center gap-3 mb-5">
            <span class="text-4xl">{{ confirmItem.icon }}</span>
            <p class="text-lg font-bold text-white">{{ confirmItem.name }}</p>

            <!-- Show preview of what it looks like -->
            <div v-if="!confirmItem.consumable" class="flex items-center gap-3 bg-dark rounded-xl p-3 w-full">
              <UserAvatar
                :name="authStore.user?.name ?? ''"
                :cosmetics="confirmPreviewCosmetics"
                :rank-top3="true"
              />
              <div>
                <UserName
                  v-if="authStore.user"
                  :name="authStore.user.name"
                  :cosmetics="confirmPreviewCosmetics"
                />
                <p class="text-xs text-gray-400">{{ confirmPreviewCosmetics?.title || 'Player' }}</p>
              </div>
            </div>

            <p v-if="confirmItem.consumable" class="text-sm text-gray-400 text-center">
              This will trigger a {{ confirmItem.name.toLowerCase() }} effect visible to everyone!
            </p>
          </div>

          <!-- Actions -->
          <div class="flex gap-3">
            <button
              class="flex-1 py-2.5 rounded-lg bg-dark-lighter text-gray-300 font-medium transition-colors hover:bg-dark"
              @click="cancelBuy"
            >
              Cancel
            </button>
            <button
              v-if="confirmMode === 'buy'"
              class="flex-1 py-2.5 rounded-lg bg-yellow-500/20 text-yellow-300 font-bold transition-colors hover:bg-yellow-500/30"
              :disabled="!canAfford(confirmItem)"
              @click="confirmBuy"
            >
              Buy · {{ confirmItem.price }} 🪙
            </button>
            <button
              v-else
              class="flex-1 py-2.5 rounded-lg font-bold transition-colors"
              :class="isEquipped(confirmItem) ? 'bg-red-500/20 text-red-300 hover:bg-red-500/30' : 'bg-primary/20 text-primary hover:bg-primary/30'"
              @click="confirmEquip"
            >
              {{ isEquipped(confirmItem) ? 'Unequip' : 'Equip' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
