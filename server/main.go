package server

import (
	"log"
	"net/http"

	"github.com/dylanmccormick/light-cycles/protocol"
	"github.com/dylanmccormick/light-cycles/server/game"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{}
	sGame    = game.CreateGame()
)

func Run() {
	go sGame.GameLoop()
	sGame.Players["player_1"] = game.Player{}
	// sGame.Players["player_2"] = game.Player{}
	http.HandleFunc("/", ping)
	http.ListenAndServe("localhost:8080", nil)
}

func ping(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
	}

	go func() {
		for {
			var input protocol.PlayerInput
			c.ReadJSON(&input)
			sGame.PlayerUpdateChan <- input
		}
	}()

	for msg := range sGame.StateUpdateChan {
		c.WriteJSON(msg)
	}
}
