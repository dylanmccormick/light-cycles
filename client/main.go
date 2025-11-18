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
			c.WriteJSON(createPlayerInput(protocol.Direction(dir), "player_1"))
		}
	}()

	select {}
}

func createPlayerInput(dir protocol.Direction, playerID string) protocol.PlayerInput {
	return protocol.PlayerInput{
		PlayerID:  playerID,
		Direction: dir,
	}
}
