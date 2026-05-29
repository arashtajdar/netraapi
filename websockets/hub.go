package websockets

import (
	"encoding/json"
	"log"
	"sync"

	"netra-api/config"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	Room   string
	UserID int
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	Rooms      map[string]map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

type Message struct {
	Room string
	Data []byte
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.Room] == nil {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true
			h.mu.Unlock()

			joinMsg, _ := json.Marshal(map[string]interface{}{"type": "userJoined"})
			h.Broadcast <- Message{Room: client.Room, Data: joinMsg}

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Rooms[client.Room][client]; ok {
				delete(h.Rooms[client.Room], client)
				close(client.Send)
				if len(h.Rooms[client.Room]) == 0 {
					delete(h.Rooms, client.Room)
				}
			}
			h.mu.Unlock()

			leftMsg, _ := json.Marshal(map[string]interface{}{"type": "userLeft"})
			h.Broadcast <- Message{Room: client.Room, Data: leftMsg}

		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Rooms[message.Room] {
				select {
				case client.Send <- message.Data:
				default:
					close(client.Send)
					delete(h.Rooms[message.Room], client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func PersistVideoState(roomCode string, isPlaying bool, positionMs int) {
	positionSec := positionMs / 1000
	query := `UPDATE watch_party_rooms SET is_playing = ?, current_position_seconds = ? WHERE room_code = ?`
	_, err := config.DB.Exec(query, isPlaying, positionSec, roomCode)
	if err != nil {
		log.Printf("Error saving room state: %v", err)
	}
}
