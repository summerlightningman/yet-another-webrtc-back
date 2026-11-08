package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type UserPayload struct {
	DisplayName string
	RoomId      RoomID
	Conn        *websocket.Conn
}

type Hub struct {
	rooms      map[RoomID]*Room
	roomByUser map[UserID]*Room
	joinRoom   chan UserPayload
	leaveRoom  chan UserID
	broadcast  chan UserMessageHubPayload
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[RoomID]*Room),
		roomByUser: make(map[UserID]*Room),
		joinRoom:   make(chan UserPayload),
		leaveRoom:  make(chan UserID),
		broadcast:  make(chan UserMessageHubPayload),
	}
}

func (h *Hub) trySend(user *User, message UserMessage) {
	select {
	case user.Send <- message:
	default:
		// Медленный клиент не должен блокировать весь hub.
		_ = user.conn.Close()
	}
}

func (h *Hub) Run() {
	for {
		select {
		case payload := <-h.joinRoom:
			room, ok := h.rooms[payload.RoomId]
			if !ok {
				room = NewRoom(payload.RoomId)
				h.rooms[payload.RoomId] = room
			}
			id := GenerateID(8, 8)
			for _, taken := h.roomByUser[id]; taken; _, taken = h.roomByUser[id] {
				id = GenerateID(8, 8)
			}
			newUser := NewUser(id, payload.DisplayName, h, payload.Conn)
			room.Join(newUser)
			h.roomByUser[newUser.Id] = room

		case message := <-h.broadcast:
			room, ok := h.roomByUser[message.User.Id]
			if !ok {
				_ = message.User.conn.Close()
				continue
			}
			switch message.Message.Type {
			case UserMessageText:
				for _, roommate := range room.Users {
					h.trySend(roommate, message.Message)
				}
				room.OnMessage(message.User.Id, message.Message)
			default:
				log.Printf("unknown message type: %v", message.Message.Type)
			}

		case userID := <-h.leaveRoom:
			room, ok := h.roomByUser[userID]
			if !ok {
				continue
			}
			user, exists := room.Users[userID]
			room.Leave(userID)
			delete(h.roomByUser, userID)
			if exists {
				close(user.Send)
			}
			if len(room.Users) == 0 {
				delete(h.rooms, room.Id)
			}
		}
	}
}
