import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useLeaderboardStore } from '@/stores/leaderboard'
import { usePredictionsStore } from '@/stores/predictions'
import { useBetsStore } from '@/stores/bets'

interface SSEEvent {
  type: 'predictions' | 'leaderboard' | 'bets'
  user_id?: string
}

export function useSSE() {
  const connected = ref(false)
  let eventSource: EventSource | null = null
  let reconnectTimeout: ReturnType<typeof setTimeout> | null = null

  function connect() {
    if (eventSource) {
      eventSource.close()
    }

    const authStore = useAuthStore()
    const url = authStore.token
      ? `/api/events?token=${encodeURIComponent(authStore.token)}`
      : '/api/events'

    eventSource = new EventSource(url)

    eventSource.onopen = () => {
      connected.value = true
      console.log('[SSE] Connected')
    }

    eventSource.onmessage = (event) => {
      try {
        const data: SSEEvent = JSON.parse(event.data)
        handleEvent(data)
      } catch (e) {
        console.error('[SSE] Failed to parse event:', e)
      }
    }

    eventSource.onerror = () => {
      connected.value = false
      console.log('[SSE] Connection error, reconnecting...')
      eventSource?.close()
      eventSource = null

      // Reconnect after 3 seconds
      reconnectTimeout = setTimeout(connect, 3000)
    }
  }

  function handleEvent(event: SSEEvent) {
    console.log('[SSE] Event:', event.type)

    const predictionsStore = usePredictionsStore()
    const betsStore = useBetsStore()
    const authStore = useAuthStore()
    const leaderboardStore = useLeaderboardStore()

    switch (event.type) {
      case 'predictions':
        // Refresh predictions list
        predictionsStore.swapPredictions()
        break

      case 'leaderboard':
        // Refresh user's token balance
        authStore.swapUser()
        leaderboardStore.swapLeaderboard()
        break

      case 'bets':
        // Refresh user's bets
        betsStore.swapBets()
        break
    }
  }

  function disconnect() {
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout)
      reconnectTimeout = null
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    connected.value = false
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    connect,
    disconnect,
  }
}
