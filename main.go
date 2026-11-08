package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

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

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		hub.joinRoom <- UserPayload{
			RoomId:      roomId,
			DisplayName: name,
			Conn:        conn,
		}
	}
}

func main() {
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/health", Health)
	http.HandleFunc("/join", HandleJoin(hub))

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
