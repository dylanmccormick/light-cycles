package game

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/dylanmccormick/light-cycles/protocol"
	"github.com/gorilla/websocket"
)

type Player struct {
	Position  protocol.Coordinate
	Direction protocol.Direction
	Trail     Queue
	Status    string
	Points    int
	Conn      *websocket.Conn
}

type GameState int

const (
	MENU GameState = iota
	COUNTDOWN
	RUNNING
	ENDED
)

type Game struct {
	Players          map[string]*Player
	PlayerUpdateChan chan (protocol.PlayerInput)
	StateUpdateChan  chan (protocol.Message)
	CommandChan      chan (protocol.GameCommand)
	tick             int
	Board            [][]int
	State            GameState
	Countdown        int
}

func CreateGame() *Game {
	return &Game{
		Players:          make(map[string]*Player),
		PlayerUpdateChan: make(chan (protocol.PlayerInput), 50),
		StateUpdateChan:  make(chan (protocol.Message)),
		CommandChan:      make(chan (protocol.GameCommand)),
		tick:             0,
	}
}

func CreatePlayerOne(points int, conn *websocket.Conn) *Player {
	return &Player{
		Trail:     Queue{protocol.TrailSegment{Coordinate: protocol.Coordinate{X: 0, Y: 15}}},
		Position:  protocol.Coordinate{X: 1, Y: 15},
		Direction: protocol.D_RIGHT,
		Status:    "alive",
		Points:    points,
		Conn: conn,
	}
}

func CreatePlayerTwo(points int, conn *websocket.Conn) *Player {
	return &Player{
		Trail:     Queue{protocol.TrailSegment{Coordinate: protocol.Coordinate{X: 46, Y: 15}}},
		Position:  protocol.Coordinate{X: 47, Y: 15},
		Direction: protocol.D_LEFT,
		Status:    "alive",
		Points:    points,
		Conn: conn,
	}
}

func (g *Game) ResetGame() {
	g.Players["player_1"] = CreatePlayerOne(g.Players["player_1"].Points, g.Players["player_1"].Conn)
	g.Players["player_2"] = CreatePlayerTwo(g.Players["player_2"].Points, g.Players["player_2"].Conn)
}

func (g *Game) ProcessCommand(cmd protocol.GameCommand) {
	if strings.Contains("start", cmd.Command) {
		g.ResetGame()
		g.State = COUNTDOWN
		return
	}
}

func (g *Game) GameHandler() {
	for {
		select {
		case cmd := <-g.CommandChan:
			g.ProcessCommand(cmd)
		default:
		}
		switch g.State {
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
			for i := countdownTime; i > 0; i-- {
				time.Sleep(1 * time.Second)
				g.StateUpdateChan <- g.buildGameState()
				g.Countdown = i
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
	var killPlayers bool
	var trailCoordinates []protocol.Coordinate
	var players []*Player
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
			player.Status = "DEAD"
			g.Players[id] = player
			killPlayers = true
		}
	}
	for i, player := range g.Players {
		if containsCoords(trailCoordinates, player.Position) {
			log.Printf("Killing player %s\n", i)
			player.Status = "DEAD"
			g.Players[i] = player
			killPlayers = true
		}
	}
	if killPlayers {
		p1 := g.Players["player_1"]
		p2 := g.Players["player_2"]
		if p1.Status == "DEAD" {
			p2.Points += 1
		}
		if p2.Status == "DEAD" {
			p1.Points += 1
		}
		g.Players["player_1"] = p1
		g.Players["player_2"] = p2
		g.State = ENDED
	}
}

func (g *Game) buildGameState() protocol.Message {
	players := make(map[string]protocol.PlayerState)
	for id, playerState := range g.Players {
		players[id] = protocol.PlayerState{
			PlayerID:  id,
			Position:  playerState.Position,
			Direction: playerState.Direction,
			Trail:     playerState.Trail,
			Status:    playerState.Status,
			Points:    playerState.Points,
		}
	}
	body, err := json.Marshal(protocol.GameState{
		Players:   players,
		Tick:      g.tick,
		Countdown: g.Countdown,
	})
	if err != nil {
		log.Fatal("Unable to marshal json for gamestate", err)
	}
	return protocol.Message{
		Type: "GameState",
		Body: body,
	}
}

func isLegalDirectionChange(original, change protocol.Direction) bool {
	switch original {
	case protocol.D_UP:
		if change == protocol.D_DOWN {
			return false
		}
	case protocol.D_DOWN:
		if change == protocol.D_UP {
			return false
		}
	case protocol.D_LEFT:
		if change == protocol.D_RIGHT {
			return false
		}
	case protocol.D_RIGHT:
		if change == protocol.D_LEFT {
			return false
		}
	}
	return true
}

func (g *Game) updatePlayer(input protocol.PlayerInput) {
	if player, exists := g.Players[input.PlayerID]; exists {
		if isLegalDirectionChange(player.Direction, input.Direction) {
			player.Direction = input.Direction
		} else {
			log.Printf("Bad direction change. Can't go from %d to %d\n", player.Direction, input.Direction)
		}
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
			player.Position.Y = 23
		}
		if player.Position.Y >= 24 {
			player.Position.Y = 0
		}
		if player.Position.X < 0 {
			player.Position.X = 47
		}
		if player.Position.X >= 48 {
			player.Position.X = 0
		}

		g.Players[id] = player
	}
}
