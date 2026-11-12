package main

type RoomID = string

type RoomMessage struct {
	Id     int    `json:"id"`
	UserID UserID `json:"userId"`
	Text   string `json:"text"`
}

type Room struct {
	Id       RoomID
	Users    map[UserID]*User
	Messages []RoomMessage
}

func NewRoom(id RoomID) *Room {
	return &Room{
		Id:    id,
		Users: make(map[UserID]*User),
	}
}

func (r *Room) Join(user *User) {
	r.Users[user.Id] = user
}

func (r *Room) Leave(userId UserID) {
	delete(r.Users, userId)
}

func (r *Room) OnMessage(userId UserID, message UserMessage) {
	messageID := 1
	if len(r.Messages) > 0 {
		messageID = r.Messages[len(r.Messages)-1].Id + 1
	}
	r.Messages = append(r.Messages, RoomMessage{
		Id:     messageID,
		UserID: userId,
		Text:   message.Payload.Text,
	})
}
