package events

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Event types
const (
	EventPredictions = "predictions" // Predictions list changed
	EventLeaderboard = "leaderboard" // Leaderboard changed (tokens changed)
	EventBets        = "bets"        // User's bets changed (for specific user)
	EventAchievement = "achievement" // User earned an achievement (for specific user)
)

// Event represents an SSE event
type Event struct {
	Type          string `json:"type"`
	UserID        string `json:"user_id,omitempty"`        // Optional: for user-specific events
	AchievementID string `json:"achievement_id,omitempty"` // Optional: for achievement events
}

// Client represents a connected SSE client
type Client struct {
	ID     string
	UserID string      // Authenticated user ID (empty if not authenticated)
	Send   chan []byte // Channel to send events
}

// Hub maintains connected clients and broadcasts events
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mu         sync.RWMutex
}

// NewHub creates a new event hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 256),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			data, _ := json.Marshal(event)
			message := fmt.Sprintf("data: %s\n\n", data)

			for client := range h.clients {
				// If event is user-specific, only send to that user
				if event.UserID != "" && client.UserID != event.UserID {
					continue
				}

				select {
				case client.Send <- []byte(message):
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Emit broadcasts an event to all clients (or specific user if UserID set)
func (h *Hub) Emit(event Event) {
	h.broadcast <- event
}

// EmitPredictions notifies all clients that predictions changed
func (h *Hub) EmitPredictions() {
	h.Emit(Event{Type: EventPredictions})
}

// EmitLeaderboard notifies all clients that leaderboard changed
func (h *Hub) EmitLeaderboard() {
	h.Emit(Event{Type: EventLeaderboard})
}

// EmitBets notifies a specific user that their bets changed
func (h *Hub) EmitBets(userID string) {
	h.Emit(Event{Type: EventBets, UserID: userID})
}

func (h *Hub) EmitBetsAll() {
	h.Emit(Event{Type: EventBets})
}

// EmitAchievement notifies a specific user that they earned an achievement
func (h *Hub) EmitAchievement(userID, achievementID string) {
	h.Emit(Event{Type: EventAchievement, UserID: userID, AchievementID: achievementID})
}
