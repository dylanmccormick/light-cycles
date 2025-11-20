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
	sGame.Players["player_1"] = CreatePlayerOne()
	sGame.Players["player_2"] = CreatePlayerTwo()
	go sGame.GameHandler()
	// go sGame.GameLoop()
	http.HandleFunc("/", ping)
	http.ListenAndServe("localhost:8080", nil)
}

func CreatePlayerOne() game.Player {
	return game.Player{
		Trail:     game.Queue{protocol.TrailSegment{Coordinate: protocol.Coordinate{X: 0, Y: 15}}},
		Position:  protocol.Coordinate{X: 1, Y: 15},
		Direction: protocol.D_RIGHT,
		Status:    "alive",
	}
}

func CreatePlayerTwo() game.Player {
	return game.Player{
		Trail:     game.Queue{protocol.TrailSegment{Coordinate: protocol.Coordinate{X: 46, Y: 15}}},
		Position:  protocol.Coordinate{X: 47, Y: 15},
		Direction: protocol.D_LEFT,
		Status:    "alive",
	}
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
