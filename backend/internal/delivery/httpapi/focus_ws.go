package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// FocusSessionEvent represents a focus session event
type FocusSessionEvent struct {
	Action    string `json:"action"` // started | stopped
	Duration  int    `json:"duration"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"user_id"`
}

// WebSocketMessage is the envelope for all WebSocket messages
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// AuthPayload is the payload for auth messages
type AuthPayload struct {
	Token string `json:"token"`
}

// Claims represents JWT claims
type FocusWSClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var jwtSecret string

// SetJWTSecret sets the JWT secret for WebSocket token validation
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

// extractUserIDFromToken extracts the user ID from a JWT token
func extractUserIDFromToken(tokenString string) (string, error) {
	if jwtSecret == "" {
		return "", errors.New("jwt secret not configured")
	}

	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &FocusWSClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*FocusWSClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token claims")
}

// FocusHub manages focus session WebSocket connections
type FocusHub struct {
	clients    map[string]*FocusClient
	mu         sync.RWMutex
	broadcast  chan *FocusSessionEvent
	register   chan *FocusClient
	unregister chan *FocusClient
}

type FocusClient struct {
	userID string
	conn   *websocket.Conn
	send   chan *FocusSessionEvent
	hub    *FocusHub
}

var focusHub *FocusHub
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	focusHub = &FocusHub{
		clients:    make(map[string]*FocusClient),
		broadcast:  make(chan *FocusSessionEvent, 256),
		register:   make(chan *FocusClient),
		unregister: make(chan *FocusClient),
	}
	go focusHub.run()
}

// GetFocusHub returns the focus hub instance
func GetFocusHub() *FocusHub {
	return focusHub
}

// HandleFocusWebSocket handles WebSocket upgrades for focus session updates
func HandleFocusWebSocket(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := &FocusClient{
		userID: userID,
		conn:   conn,
		send:   make(chan *FocusSessionEvent, 256),
		hub:    focusHub,
	}

	focusHub.register <- client

	go client.readPump()
	go client.writePump()
}

// BroadcastFocusStarted broadcasts a focus session start event
func (h *FocusHub) BroadcastFocusStarted(userID string, duration int) {
	event := &FocusSessionEvent{
		UserID:    userID,
		Action:    "started",
		Duration:  duration,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcast <- event
}

// BroadcastFocusStopped broadcasts a focus session stop event
func (h *FocusHub) BroadcastFocusStopped(userID string) {
	event := &FocusSessionEvent{
		UserID:    userID,
		Action:    "stopped",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcast <- event
}

func (h *FocusHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			slog.Info("focus client registered", "user_id", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if c, ok := h.clients[client.userID]; ok && c == client {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Info("focus client unregistered", "user_id", client.userID)

		case event := <-h.broadcast:
			h.mu.RLock()
			// Send to the specific user's client if connected
			if client, ok := h.clients[event.UserID]; ok {
				select {
				case client.send <- event:
				default:
					// Channel full, drop event
					slog.Warn("focus event dropped", "user_id", event.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *FocusClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// First message should be auth if userID not set from JWT
	if c.userID == "" {
		var msg WebSocketMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
				slog.Error("websocket read failed", "error", err)
			return
		}

		if msg.Type != "auth" {
			slog.Warn("expected auth message", "type", msg.Type)
			c.conn.WriteJSON(gin.H{"error": "authentication required"})
			return
		}

		// Parse auth payload
		payloadMap, ok := msg.Payload.(map[string]interface{})
		if !ok {
			slog.Warn("invalid auth payload")
			c.conn.WriteJSON(gin.H{"error": "invalid auth payload"})
			return
		}

		token, ok := payloadMap["token"].(string)
		if !ok || token == "" {
			slog.Warn("missing token in auth payload")
			c.conn.WriteJSON(gin.H{"error": "missing token"})
			return
		}

		// Validate token and extract userID
		userID, err := extractUserIDFromToken(token)
		if err != nil {
			slog.Error("invalid token", "error", err)
			c.conn.WriteJSON(gin.H{"error": "invalid token"})
			return
		}

		c.userID = userID
		slog.Info("focus client authenticated", "user_id", c.userID)
	}

	for {
		var msg WebSocketMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket error", "error", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "focus_session_event":
			// Parse focus event
			payloadMap, ok := msg.Payload.(map[string]interface{})
			if !ok {
				slog.Warn("invalid focus_session_event payload")
				continue
			}

			event := &FocusSessionEvent{
				UserID:    c.userID,
				Action:    "started",
				Duration:  0,
				Timestamp: time.Now().Format(time.RFC3339),
			}

			if action, ok := payloadMap["action"].(string); ok {
				if action == "start" {
					event.Action = "started"
				} else if action == "stop" {
					event.Action = "stopped"
				}
			}

			if duration, ok := payloadMap["duration"].(float64); ok {
				event.Duration = int(duration)
			}

			c.hub.broadcast <- event
		}
	}
}

func (c *FocusClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			msg := WebSocketMessage{
				Type:    "focus_session_event",
				Payload: event,
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
