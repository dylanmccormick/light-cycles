package game

import (
	"log"
	"time"

	"github.com/dylanmccormick/light-cycles/protocol"
)

type Player struct {
	PosX      int
	PosY      int
	Direction protocol.Direction
}

type Game struct {
	Players          map[string]Player
	PlayerUpdateChan chan (protocol.PlayerInput)
	StateUpdateChan  chan (protocol.GameState)
	tick             int
}

func CreateGame() *Game {
	return &Game{
		Players:          make(map[string]Player),
		PlayerUpdateChan: make(chan (protocol.PlayerInput), 50),
		StateUpdateChan:  make(chan (protocol.GameState)),
		tick:             0,
	}
}

func (g *Game) GameLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	done := make(chan bool)
	for {
		select {
		case <-done:
			log.Println("DONE")
		case <-ticker.C:
			g.moveAllPlayers()
			g.StateUpdateChan <- g.buildGameState()
			g.tick += 1
		case input := <-g.PlayerUpdateChan:
			g.updatePlayer(input)
		}
	}
}

func (g *Game) buildGameState() protocol.GameState {
	players := make(map[string]protocol.PlayerState)
	for id, state := range g.Players {
		players[id] = protocol.PlayerState{
			PlayerID:  id,
			PosX:      state.PosX,
			PosY:      state.PosY,
			Direction: state.Direction,
		}
	}
	return protocol.GameState{
		Players: players,
		Tick:    g.tick,
	}
}

func (g *Game) updatePlayer(input protocol.PlayerInput) {
	if player, exists := g.Players[input.PlayerID]; exists {
		log.Printf("Updating %s's direction to %d\n", input.PlayerID, input.Direction)
		player.Direction = input.Direction
		g.Players[input.PlayerID] = player
	}
}

func (g *Game) moveAllPlayers() {
	for id, player := range g.Players {
		log.Printf("Updating %s's position", id)
		switch player.Direction {
		case protocol.D_UP:
			player.PosY -= 1
		case protocol.D_DOWN:
			player.PosY += 1
		case protocol.D_LEFT:
			player.PosX -= 1
		case protocol.D_RIGHT:
			player.PosX += 1
		}

		g.Players[id] = player
	}
}
