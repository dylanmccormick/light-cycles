package protocol

type Direction int

const (
	D_UP Direction = iota
	D_DOWN
	D_LEFT
	D_RIGHT
)

type PlayerState struct {
	PlayerID  string    `json:"player_id"`
	PosX      int       `json:"pos_x"`
	PosY      int       `json:"pos_y"`
	Direction Direction `json:"facing"`
}
type GameState struct {
	Players map[string]PlayerState `json:"players"`
	Tick    int                    `json:"tick"`
}

type PlayerInput struct {
	PlayerID  string    `json:"player_id"`
	Direction Direction `json:"direction"`
}
