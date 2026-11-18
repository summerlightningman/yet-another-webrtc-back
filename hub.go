package main

import (
	"fmt"
	"log"
	"time"

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
		rooms:      make(map[RoomID]*Room, 100),
		roomByUser: make(map[UserID]*Room, 100),
		joinRoom:   make(chan UserPayload, 100),
		leaveRoom:  make(chan UserID, 100),
		broadcast:  make(chan UserMessageHubPayload, 100),
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

func (h *Hub) HandleJoinRoom(payload UserPayload) {
	log.Printf("Join room RoomID=%s, DisplayName=%s, RoomID=%s\n", payload.RoomId, payload.DisplayName, payload.RoomId)
	room, ok := h.rooms[payload.RoomId]
	if !ok {
		room = NewRoom(payload.RoomId)
		h.rooms[payload.RoomId] = room
		log.Printf("New room with ID=%s\n", room.Id)
	}
	id := GenerateID(8, 8)
	for _, taken := h.roomByUser[id]; taken; _, taken = h.roomByUser[id] {
		id = GenerateID(8, 8)
	}
	newUser := NewUser(id, payload.DisplayName, h, payload.Conn)
	log.Printf("New User=%s DisplayName=%s\n", newUser.Id, newUser.DisplayName)
	room.Join(newUser)
	h.roomByUser[newUser.Id] = room

	roommates := make([]Roommate, 0, len(room.Users))
	for _, roommate := range room.Users {
		roommates = append(roommates, Roommate{
			Id:          roommate.Id,
			DisplayName: roommate.DisplayName,
		})
	}

	messages := make([]RoomMessage, 0, len(room.Messages))
	for _, msg := range room.Messages {
		messages = append(messages, msg)
	}

	h.trySend(newUser, UserMessage{
		Type: Introduction,
		Payload: UserMessagePayload{
			Roommates: roommates,
			Messages:  messages,
			You: Roommate{
				Id:          newUser.Id,
				DisplayName: newUser.DisplayName,
			},
		},
	})
	log.Printf("Send to broadcast introduction to %s\n", newUser.DisplayName)
	h.broadcast <- UserMessageHubPayload{
		Message: UserMessage{
			Type: RoomUpdate,
			Payload: UserMessagePayload{
				Text: fmt.Sprintf("Пользователь %s вошел в комнату", newUser.DisplayName),
			},
		},
		User:      newUser,
		Timestamp: time.Now(),
	}
}
func (h *Hub) HandleLeaveRoom(userID UserID) {
	room, ok := h.roomByUser[userID]
	if !ok {
		return
	}
	user, exists := room.Users[userID]
	room.Leave(userID)
	h.broadcast <- UserMessageHubPayload{
		Message: UserMessage{
			Type: LeaveRoom,
			Payload: UserMessagePayload{
				Text: fmt.Sprintf("Пользователь %s покинул комнату", user.DisplayName),
			},
		},

		Timestamp: time.Now(),
	}
	delete(h.roomByUser, userID)
	if exists {
		close(user.Send)
	}
	if len(room.Users) == 0 {
		delete(h.rooms, room.Id)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case payload := <-h.joinRoom:
			h.HandleJoinRoom(payload)

		case message := <-h.broadcast:
			room, ok := h.roomByUser[message.User.Id]
			if !ok {
				_ = message.User.conn.Close()
				continue
			}
			switch message.Message.Type {
			case Text:
				for _, roommate := range room.Users {
					h.trySend(roommate, message.Message)
				}
				room.OnMessage(message.User.Id, message.Message)
			case Introduction:
				fallthrough
			case LeaveRoom:
				fallthrough
			case RoomUpdate:
				log.Printf("Making broadcast %d of %s\n", message.Message.Type, message.User.Id)

				roommates := make([]Roommate, 0, len(room.Users))
				for _, roommate := range room.Users {
					roommates = append(roommates, Roommate{
						Id:          roommate.Id,
						DisplayName: roommate.DisplayName,
					})
				}

				for _, roommate := range room.Users {
					if message.User.Id == roommate.Id {
						continue
					}
					log.Printf("Send roomstate of %s to %s\n", message.User.Id, roommate.DisplayName)
					h.trySend(roommate, UserMessage{
						Type: message.Message.Type,
						Payload: UserMessagePayload{
							Roommates: roommates,
							Text:      message.Message.Payload.Text,
						},
					})
				}
			}
		case userID := <-h.leaveRoom:
			h.HandleLeaveRoom(userID)
		}
	}
}
