package main

type UserPayload struct {
	DisplayName string
	RoomId      RoomID
}

type Hub struct {
	rooms      map[RoomID]*Room
	unregister chan RoomID
	joinRoom   chan UserPayload
	leaveRoom  chan UserID
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[RoomID]*Room),
		unregister: make(chan RoomID),
		joinRoom:   make(chan UserPayload),
		leaveRoom:  make(chan UserID),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case roomId := <-h.unregister:
			if _, ok := h.rooms[roomId]; ok {
				delete(h.rooms, roomId)
			}
		case payload := <-h.joinRoom:
			room, ok := h.rooms[payload.RoomId]
			if !ok {
				room = &Room{Id: payload.RoomId}
			}
			newUser := NewUser("", payload.DisplayName)
			room.Join(newUser)
		case userId := <-h.leaveRoom:

		}
	}
}
