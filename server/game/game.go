package game

import (
	"log"
	"time"

	"github.com/dylanmccormick/light-cycles/protocol"
)

type Player struct {
	Position  protocol.Coordinate
	Direction protocol.Direction
	Trail     Queue
}

type Game struct {
	Players          map[string]Player
	PlayerUpdateChan chan (protocol.PlayerInput)
	StateUpdateChan  chan (protocol.GameState)
	tick             int
	Board            [][]int
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
	for id, playerState := range g.Players {
		players[id] = protocol.PlayerState{
			PlayerID:  id,
			Position:  playerState.Position,
			Direction: playerState.Direction,
			Trail:     playerState.Trail,
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
		player.Trail.Enqueue(player.Position)
		if len(player.Trail) > 10 {
			player.Trail.Dequeue()
		}

		switch player.Direction {
		case protocol.D_UP:
			player.Position.Y -= 1
		case protocol.D_DOWN:
			player.Position.Y += 1
		case protocol.D_LEFT:
			player.Position.X -= 1
		case protocol.D_RIGHT:
			player.Position.X += 1
		}

		if player.Position.Y < 0 {
			log.Printf("looping off top")
			player.Position.Y = 23
		}
		if player.Position.Y >= 24 {
			log.Printf("looping off bottom")
			player.Position.Y = 0
		}
		if player.Position.X < 0 {
			log.Printf("looping off left")
			player.Position.X = 23
		}
		if player.Position.X >= 24 {
			log.Printf("looping off right")
			player.Position.X = 0
		}

		g.Players[id] = player
	}
}
