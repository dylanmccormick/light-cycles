package server

import (
	"encoding/json"
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
	sGame.Players["player_1"] = game.CreatePlayerOne(0)
	sGame.Players["player_2"] = game.CreatePlayerTwo(0)
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
			processMessage(c)
			// var input protocol.PlayerInput
			// c.ReadJSON(&input)
			// sGame.PlayerUpdateChan <- input
		}
	}()

	for msg := range sGame.StateUpdateChan {
		c.WriteJSON(msg)
	}
}

func processMessage(c *websocket.Conn) {
	var input protocol.Message
	c.ReadJSON(&input)
	switch input.Type {
	case "GameCommand":
		var gc protocol.GameCommand
		err := json.Unmarshal(input.Body, &gc)
		if err != nil {
			log.Fatal("Unable to unmarshal gameCommand:", err)
		}
		sGame.CommandChan <- gc
	case "PlayerInput":
		var pi protocol.PlayerInput
		err := json.Unmarshal(input.Body, &pi)
		if err != nil {
			log.Fatal("Unable to unmarshal gameCommand:", err)
		}
		sGame.PlayerUpdateChan <- pi
	}
}
