package protocol

type Direction int

const (
	D_UP Direction = iota
	D_DOWN
	D_LEFT
	D_RIGHT
)

type PlayerState struct {
	PlayerID  string       `json:"player_id"`
	Position  Coordinate   `json:"position"`
	Direction Direction    `json:"facing"`
	Trail     []Coordinate `json:"trail"`
}
type GameState struct {
	Players map[string]PlayerState `json:"players"`
	Tick    int                    `json:"tick"`
}

type PlayerInput struct {
	PlayerID  string    `json:"player_id"`
	Direction Direction `json:"direction"`
}

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}
