package protocol

import "encoding/json"

type Direction int

const (
	D_UP Direction = iota
	D_DOWN
	D_LEFT
	D_RIGHT
)

type Message struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

type PlayerAssignment struct {
	PlayerID string `json:"player_id"`
}
type PlayerState struct {
	PlayerID  string         `json:"player_id"`
	Position  Coordinate     `json:"position"`
	Direction Direction      `json:"direction"`
	Trail     []TrailSegment `json:"trail"`
	Status    string         `json:"status"`
	Points    int            `json:"points"`
}

type TrailSegment struct {
	Coordinate Coordinate `json:"position"`
	Direction  Direction  `json:"direction"`
}

type GameCommand struct {
	Command string `json:"command"`
}

type GameState struct {
	Players   map[string]PlayerState `json:"players"`
	Tick      int                    `json:"tick"`
	Countdown int                    `json:"countdown"` // time until game starts
}

type PlayerInput struct {
	PlayerID  string    `json:"player_id"`
	Direction Direction `json:"direction"`
}

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}
