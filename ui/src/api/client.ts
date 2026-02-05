import type { User, LeaderboardUser } from '@/types/users'
import type { Prediction, PredictionWithOdds, Bet } from '@/types/predictions'

const API_BASE = '/api'

interface ApiError {
  error: string
}

class ApiClient {
  private token: string | null = null

  setToken(token: string | null) {
    this.token = token
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    if (response.status === 204) {
      return undefined as T
    }

    return response.json()
  }

  // Auth endpoints
  async register(name: string, pin: string): Promise<{ token: string; user: User }> {
    return this.request('POST', '/register', { name, pin })
  }

  async login(name: string, pin: string): Promise<{ token: string; user: User }> {
    return this.request('POST', '/login', { name, pin })
  }

  async getMe(): Promise<User> {
    return this.request('GET', '/me')
  }

  // Predictions endpoints
  async getPredictions(): Promise<PredictionWithOdds[]> {
    return this.request('GET', '/predictions')
  }

  async getPrediction(id: string): Promise<PredictionWithOdds> {
    return this.request('GET', `/predictions/${id}`)
  }

  async placeBet(predictionId: string, choiceId: string, amount: number): Promise<Bet> {
    // todo: needs to support updates properly
    return this.request('POST', '/bets', {
      prediction_id: predictionId,
      prediction_choice_id: choiceId,
      amount,
    })
  }

  // User endpoints
  async getMyBets(): Promise<Bet[]> {
    return this.request('GET', '/my-bets')
  }

  async getLeaderboard(): Promise<LeaderboardUser[]> {
    return this.request('GET', '/leaderboard')
  }

  // Admin endpoints
  async createPrediction(data: {
    name: string
    description: string
    closes_at: string
    choices: { name: string }[]
    odds_visible_before_bet: boolean
  }): Promise<Prediction> {
    return this.request('POST', '/admin/predictions', data)
  }

  async updatePrediction(id: string, data: {
    name?: string
    description?: string
    closes_at?: string
    odds_visible_before_bet?: boolean
  }): Promise<Prediction> {
    return this.request('PUT', `/admin/predictions/${id}`, data)
  }

  async closePrediction(id: string): Promise<Prediction> {
    return this.request('POST', `/admin/predictions/${id}/close`)
  }

  async decidePrediction(id: string, winningChoiceId: string): Promise<Prediction> {
    return this.request('POST', `/admin/predictions/${id}/decide`, {
      winning_choice_id: winningChoiceId,
    })
  }

  async voidPrediction(id: string): Promise<Prediction> {
    return this.request('POST', `/admin/predictions/${id}/void`)
  }

  async getUsers(): Promise<User[]> {
    return this.request('GET', '/admin/users')
  }

  async giftTokens(userId: string, amount: number): Promise<User> {
    return this.request('POST', `/admin/users/${userId}/tokens`, { amount })
  }

  async resetUserPin(userId: string, newPin: string): Promise<void> {
    return this.request('POST', `/admin/users/${userId}/reset-pin`, { pin: newPin })
  }
}

export const api = new ApiClient()
