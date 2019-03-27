package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// upgrade original http request to ws
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go func(conn *websocket.Conn) {
			for {
				mType, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					return
				}

				conn.WriteMessage(mType, msg)
			}
		}(conn)
	})

	http.HandleFunc("/v2/ws", func(w http.ResponseWriter, r *http.Request) {
		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go func(conn *websocket.Conn) {
			for {
				_, msg, _ := conn.ReadMessage()
				println(string(msg))
			}
		}(conn)
	})

	// every json every 5 minutes to channel
	http.HandleFunc("/v3/ws", func(w http.ResponseWriter, r *http.Request) {
		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go func(conn *websocket.Conn) {
			ch := time.Tick(5 * time.Second)

			for range ch {
				conn.WriteJSON(myStruct{
					Username:  "cristobal",
					FirstName: "chris",
					LastName:  "Ffffff",
				})
			}
		}(conn)
	})

	http.ListenAndServe(":3000", nil)
}

type myStruct struct {
	Username  string `json:"username"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
}
