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
	Status    string
}

type GameState int

const (
	MENU GameState = iota
	COUNTDOWN
	RUNNING
	ENDED
)

type Game struct {
	Players          map[string]Player
	PlayerUpdateChan chan (protocol.PlayerInput)
	StateUpdateChan  chan (protocol.GameState)
	CommandChan      chan (protocol.GameCommand)
	tick             int
	Board            [][]int
	State            GameState
}

func CreateGame() *Game {
	return &Game{
		Players:          make(map[string]Player),
		PlayerUpdateChan: make(chan (protocol.PlayerInput), 50),
		StateUpdateChan:  make(chan (protocol.GameState)),
		CommandChan:      make(chan (protocol.GameCommand)),
		tick:             0,
	}
}

func (g *Game) GameHandler() {
	for {
		switch g.State {
		case MENU:
			time.Sleep(5 * time.Second)
			g.State = COUNTDOWN
		case COUNTDOWN:
			g.GameLoop()
		}
	}
}

func (g *Game) GameLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)
	countdownTime := 10
	for {
		if g.State == COUNTDOWN {
			for i := range countdownTime {
				log.Println("Time to start:", i)
				time.Sleep(1 * time.Second)
				continue
			}
			g.State = RUNNING
		}

		if g.State == ENDED {
			break
		}
		select {
		case <-done:
			log.Println("DONE")
		case <-ticker.C:
			g.moveAllPlayers()
			g.checkForCollisions()
			g.StateUpdateChan <- g.buildGameState()
			g.tick += 1
		case input := <-g.PlayerUpdateChan:
			g.updatePlayer(input)
		case input := <-g.CommandChan:
			g.executeCommand(input)
		}
	}
}

func (g *Game) executeCommand(command protocol.GameCommand) {
	switch command.Command {
	case "start":
		g.State = RUNNING
	}
}

func containsCoords(coords []protocol.Coordinate, target protocol.Coordinate) bool {
	for _, c := range coords {
		if c.X == target.X && c.Y == target.Y {
			return true
		}
	}
	return false
}

func (g *Game) checkForCollisions() {
	var trailCoordinates []protocol.Coordinate
	var players []Player
	for _, player := range g.Players {
		players = append(players, player)
		for i, ts := range player.Trail {
			// skip "ghost" coordinate
			if i == 0 {
				continue
			}
			trailCoordinates = append(trailCoordinates, ts.Coordinate)
		}
	}
	if players[0].Position == players[1].Position {
		for id, player := range g.Players {
			log.Println("COLLIDED IN SAME SPOT")
			player.Status = "DEAD"
			g.Players[id] = player
		}
	}
	log.Println("trails:", trailCoordinates)
	for i, player := range g.Players {
		log.Println("player pos", player.Position)
		kill := containsCoords(trailCoordinates, player.Position)
		log.Println("SHOULD KILL?:", kill)
		if containsCoords(trailCoordinates, player.Position) {
			log.Printf("Killing player %s\n", i)
			player.Status = "DEAD"
			g.Players[i] = player
			g.State = ENDED
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
			Status:    playerState.Status,
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
		if player.Status == "DEAD" {
			continue
		}
		player.Trail.Enqueue(protocol.TrailSegment{Coordinate: player.Position, Direction: player.Direction})
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
			player.Position.X = 47
		}
		if player.Position.X >= 48 {
			log.Printf("looping off right")
			player.Position.X = 0
		}

		g.Players[id] = player
	}
}
