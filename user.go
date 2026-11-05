package main

import "github.com/gorilla/websocket"

type UserID = string

type User struct {
	Id          UserID `json:"id"`
	DisplayName string `json:"name"`
	hub         *Hub
	conn        *websocket.Conn
}

func (u *User) ReadPump() {
	defer func() {
		u.hub.leaveRoom <- u.Id
		u.conn.Close()
	}()

}

func (u *User) WritePump() {

}

func NewUser(id UserID, displayName string) *User {
	return &User{
		Id:          id,
		DisplayName: displayName,
	}
}
