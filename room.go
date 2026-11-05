package main

import "strconv"

type RoomID = string

type Room struct {
	Id    RoomID
	Users map[UserID]*User
}

func (r *Room) Join(user *User) {
	user.Id = strconv.Itoa(len(r.Users) + 1)
	r.Users[r.Id] = user
}

func (r *Room) Leave(userId UserID) {
	if _, exists := r.Users[userId]; exists {
		delete(r.Users, userId)
	}
}
