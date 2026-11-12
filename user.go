package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type UserID = string

type UserMessageType int

const (
	Text UserMessageType = iota + 1
	Introduction
)

const (
	writeWait      = 10 * time.Second
	sendBufferSize = 64
)

type UserMessagePayload struct {
	Roommates []Roommate    `json:"roommates"`
	Messages  []RoomMessage `json:"messages"`
	You       Roommate      `json:"you"`
	Text      string        `json:"text"`
}

type UserMessage struct {
	Type    UserMessageType    `json:"type"`
	Payload UserMessagePayload `json:"payload"`
}

type Roommate struct {
	Id          UserID `json:"id"`
	DisplayName string `json:"name"`
}

type UserMessageHubPayload struct {
	Message   UserMessage
	User      *User
	Timestamp time.Time
}

type User struct {
	Id          UserID `json:"id"`
	DisplayName string `json:"name"`
	hub         *Hub
	conn        *websocket.Conn
	Send        chan UserMessage
}

func (u *User) ReadPump() {
	defer func() {
		u.hub.leaveRoom <- u.Id
		_ = u.conn.Close()
	}()

	for {
		var message UserMessage
		err := u.conn.ReadJSON(&message)
		if err != nil {
			log.Println("read error:", err)
			return
		}

		switch message.Type {
		case Text:
			u.hub.broadcast <- UserMessageHubPayload{
				Message:   message,
				User:      u,
				Timestamp: time.Now(),
			}
		default:
			log.Printf("user %s: unknown message type %v", u.Id, message.Type)
		}
	}
}

func (u *User) WritePump() {
	defer u.conn.Close()

	for message := range u.Send {
		_ = u.conn.SetWriteDeadline(time.Now().Add(writeWait))

		if err := u.conn.WriteJSON(message); err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

func NewUser(id UserID, displayName string, hub *Hub, conn *websocket.Conn) *User {
	user := &User{
		Id:          id,
		DisplayName: displayName,
		hub:         hub,
		conn:        conn,
		Send:        make(chan UserMessage, sendBufferSize),
	}

	go user.ReadPump()
	go user.WritePump()

	return user
}
