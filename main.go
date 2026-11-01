package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type User struct {
	Id          int    `json:"id"`
	DisplayName string `json:"name"`
	Description string `json:"description"`
}

type RoomUser struct {
	User
	Conn *websocket.Conn
	Room *Room
}

type Room struct {
	Users []*RoomUser
}

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

func Hub(user *RoomUser, c *websocket.Conn) {
	defer c.Close()
	roommates := make([]*User, len(user.Room.Users))
	for idx := range user.Room.Users {
		roommates[idx] = &user.Room.Users[idx].User
	}

	message := HubMessage{
		MsgType: MessageTypeRoomState,
		Payload: HubMessageJoinRoom{
			Roommates: roommates,
			You:       &user.User,
		},
	}
	err := c.WriteJSON(&message)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		message := UserMessage{}
		// Если когда-нибудь тут будет большой RPS: заменить на
		// ReadMessage|WriteMessage с easyjson или аналогом :)
		err := c.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(message)

		// err = c.WriteJSON(&message)
		// if err != nil {
		// 	log.Println(err)
		// 	break
		// }
	}
}

func Signaling(roomsMap map[string]*Room) func(http.ResponseWriter, *http.Request) {
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

		room, isRoomExists := roomsMap[roomId]

		if isRoomExists {
			newUser := RoomUser{Conn: c, Room: room, User: User{Id: len(room.Users) + 1, DisplayName: name}}
			room.Users = append(room.Users, &newUser)
			Hub(&newUser, c)
		} else {
			newRoom := Room{Users: make([]*RoomUser, 0, 50)}
			newUser := RoomUser{User: User{Id: 1, DisplayName: name}, Conn: c, Room: &newRoom}
			newRoom.Users = append(newRoom.Users, &newUser)
			roomsMap[roomId] = &newRoom
			Hub(&newUser, c)
		}
	}
}

func main() {
	roomsMap := make(map[string]*Room)
	http.HandleFunc("/health", Health)
	http.HandleFunc("/join", Signaling(roomsMap))

	http.ListenAndServe("0.0.0.0:8080", nil)
}
