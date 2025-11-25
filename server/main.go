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
	go sGame.GameHandler()
	go broadcastMessages()
	http.HandleFunc("/", ping)
	http.ListenAndServe("localhost:8080", nil)
}

func CreatePlayerOne(c *websocket.Conn) *game.Player {
	return &game.Player{
		Trail:     game.Queue{protocol.TrailSegment{Coordinate: protocol.Coordinate{X: 0, Y: 15}}},
		Position:  protocol.Coordinate{X: 1, Y: 15},
		Direction: protocol.D_RIGHT,
		Status:    "alive",
		Conn:      c,
		Points:    0,
	}
}

func CreatePlayerTwo(c *websocket.Conn) *game.Player {
	return &game.Player{
		Trail:     game.Queue{protocol.TrailSegment{Coordinate: protocol.Coordinate{X: 46, Y: 15}}},
		Position:  protocol.Coordinate{X: 47, Y: 15},
		Direction: protocol.D_LEFT,
		Status:    "alive",
		Conn:      c,
		Points:    0,
	}
}

func sendPlayerAssignmentMessage(c *websocket.Conn, playerID string) {
	body, err := json.Marshal(protocol.PlayerAssignment{
		PlayerID: playerID,
	})
	if err != nil {
		log.Fatal(err)
	}
	msg := protocol.Message{
		Type: "PlayerAssignment",
		Body: body,
	}
	c.WriteJSON(msg)
}

func ping(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
	}
	var playerAssignment string
	if _, ok := sGame.Players["player_1"]; !ok {
		playerAssignment = "player_1"
		sGame.Players["player_1"] = CreatePlayerOne(c)
	} else {
		playerAssignment = "player_2"
		sGame.Players["player_2"] = CreatePlayerTwo(c)
	}

	sendPlayerAssignmentMessage(c, playerAssignment)

	go func() {
		for {
			processMessage(c)
		}
	}()

	select {}
}

func broadcastMessages() {
	for msg := range sGame.StateUpdateChan {
		for _, player := range sGame.Players {
			if player.Conn != nil {
				player.Conn.WriteJSON(msg)
			}
		}
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
