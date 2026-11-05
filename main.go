package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type MessageType int

const (
	MessageTypeLeaveRoom MessageType = iota
	MessageTypeJoinRoom
	MessageTypeRoomState
	MessageTypeText
)

type HubMessage struct {
	MsgType MessageType `json:"type"`
	Payload any         `json:"payload"`
}

type HubMessageRoomState struct {
	Roommates *User `json:"roommates"`
}

type HubMessageJoinRoom struct {
	Roommates []*User `json:"roommates"`
	You       *User   `json:"you"`
}

type UserMessage struct {
	MsgType MessageType `json:"type"`
	Payload string      `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'm healthy!"))
}

func HandleJoin(hub *Hub) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		roomId := r.URL.Query().Get("roomId")
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if roomId == "" {
			http.Error(w, "room id is required", http.StatusBadRequest)
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
		}

	}
}

func main() {
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/health", Health)
	http.HandleFunc("/join", HandleJoin(hub))

	http.ListenAndServe("0.0.0.0:8080", nil)
}
