package websockets

import (
	"encoding/json"
	"log"
	"net/http"

	"sheedbox-api/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		tokenStr := r.URL.Query().Get("token")

		if room == "" {
			http.Error(w, "Room code required", http.StatusBadRequest)
			return
		}

		var userID int
		if tokenStr != "" {
			// Use centralized key function that enforces HS256
			token, err := jwt.Parse(tokenStr, config.JWTKeyFunc)
			if err == nil && token.Valid {
				claims := token.Claims.(jwt.MapClaims)
				if uid, ok := claims["user_id"].(float64); ok {
					userID = int(uid)
				}
			}
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		client := &Client{
			Hub:    hub,
			Room:   room,
			UserID: userID,
			Conn:   conn,
			Send:   make(chan []byte, 256),
		}

		client.Hub.Register <- client

		go client.writePump()
		go client.readPump()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS error: %v", err)
			}
			break
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(message, &payload); err == nil {
			msgType, _ := payload["type"].(string)

			switch msgType {
			case "play":
				PersistVideoState(c.Room, true, 0)
				c.Hub.Broadcast <- Message{Room: c.Room, Data: message}
			case "pause":
				PersistVideoState(c.Room, false, 0)
				c.Hub.Broadcast <- Message{Room: c.Room, Data: message}
			case "seek":
				posMs, _ := payload["positionMs"].(float64)
				PersistVideoState(c.Room, true, int(posMs))
				c.Hub.Broadcast <- Message{Room: c.Room, Data: message}
			case "newMessage":
				c.Hub.Broadcast <- Message{Room: c.Room, Data: message}
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
