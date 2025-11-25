package client

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/url"
	"time"

	"github.com/dylanmccormick/light-cycles/protocol"
	"github.com/gorilla/websocket"
)

func Connect() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			gs := &protocol.GameState{}
			json.Unmarshal(message, gs)
			log.Printf("%#v\n", gs)
		}
	}()

	go func() {
		for {
			sleepTime := rand.IntN(10)
			time.Sleep(time.Duration(sleepTime) * time.Second)
			dir := rand.IntN(4)
			c.WriteJSON(createPlayerMessage(protocol.Direction(dir), "player_1"))
		}
	}()

	select {}
}

func createPlayerMessage(dir protocol.Direction, playerID string) protocol.Message {
	pi := protocol.PlayerInput{
		PlayerID:  playerID,
		Direction: dir,
	}
	body, err := json.Marshal(pi)
	if err != nil {
		log.Fatal("You did a bad json")
	}
	return protocol.Message{
		Type: "PlayerInput",
		Body: body,
	}
}

func createGameCommand(cmd string) protocol.Message {
	gc := protocol.GameCommand{
		Command: cmd,
	}
	body, err := json.Marshal(gc)
	if err != nil {
		log.Fatal("You did a bad json")
	}
	return protocol.Message{
		Type: "GameCommand",
		Body: body,
	}
}
