package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type User struct {
	displayName string
	conn *websocket.Conn
}

type Room struct {
	users []User
}

type Message struct {
	msgType string
	text string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func (r *http.Request) bool {
		return true
	},
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'm healthy!"))
}

func Hub(user *User, c *websocket.Conn) {
	defer c.Close()
	for {
		message := Message {}
		// Если когда-нибудь тут будет большой RPS: заменить на
		// ReadMessage|WriteMessage с easyjson или аналогом :)
		err := c.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
		err = c.WriteJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func Signaling(roomsMap map[string]*Room) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		roomId := r.PathValue("id")
		name := r.PathValue("name")

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
		}


		room, isRoomExists := roomsMap[roomId]

		newUser := User { displayName: name, conn: c }
		if isRoomExists {
			room.users = append(room.users, newUser)
		} else {
			roomsMap[roomId] = &Room { users: []User{ newUser } }
		}

		Hub(&newUser, c)
	}
}

func main() {
	roomsMap := make(map[string]*Room)
	http.HandleFunc("/health", Health)
	http.HandleFunc("/join/{id}/{name}", Signaling(roomsMap))

	http.ListenAndServe("0.0.0.0:8080", nil)
}
