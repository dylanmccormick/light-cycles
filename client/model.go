package client

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/light-cycles/protocol"
	"github.com/gorilla/websocket"
)

type RootModel struct {
	width  int
	height int

	Conn     *websocket.Conn
	GameComp *GameComponent
	Messages chan (protocol.GameState)
}

type TrailRune rune

const (
	UP_RIGHT   TrailRune = '┌'
	UP_LEFT    TrailRune = '┐'
	DOWN_RIGHT TrailRune = '└'
	DOWN_LEFT  TrailRune = '┘'
	LEFT_UP    TrailRune = '└'
	LEFT_DOWN  TrailRune = '┌'
	RIGHT_UP   TrailRune = '┘'
	RIGHT_DOWN TrailRune = '┐'
	HORIZONTAL TrailRune = '─'
	VERTICAL   TrailRune = '│'
)

var clearedBoard [][]rune

type GameComponent struct {
	Players map[string]protocol.PlayerState
	Board   [][]rune
	Tick    int
	Conn    *websocket.Conn
}

func NewGameComponent(conn *websocket.Conn) *GameComponent {
	return &GameComponent{
		Players: make(map[string]protocol.PlayerState),
		Board:   clearedBoard,
		Conn:    conn,
	}
}

func (g *GameComponent) View() string {
	boardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))
	s := fmt.Sprintf("tick: %d\n", g.Tick)
	s += fmt.Sprintf("player1: %d, %d, %s\n", g.Players["player_1"].Position.X, g.Players["player_1"].Position.Y, g.Players["player_1"].Status)
	return s + boardStyle.Render(boardToString(g.Board))
}

func (g *GameComponent) Update(msg tea.Msg) (GameComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case GameStateMsg:
		g.drawBoard(msg)
		return *g, pollGame(g.Conn)
	}
	return *g, nil
}

func boardToString(board [][]rune) string {
	var out strings.Builder

	for i := range board {
		for _, char := range board[i] {
			out.WriteString(string(char))
		}
		out.WriteString("\n")
	}
	return out.String()
}

func (g *GameComponent) drawBoard(msg GameStateMsg) {
	g.Players = msg.Message.Players
	g.Tick = msg.Message.Tick
	board := copyBoard(clearedBoard)
	for _, player := range g.Players {
		px := player.Position.X
		py := player.Position.Y
		board[py][px] = getDirectionalPlayerRune(player.Direction)
		for i, trailSeg := range player.Trail {
			if i == 0 {
				continue
			}
			tx := trailSeg.Coordinate.X
			ty := trailSeg.Coordinate.Y
			board[ty][tx] = rune(getDirectionalTrailRune(trailSeg, player.Trail[i-1]))
		}
	}
	g.Board = board
}

func getDirectionalTrailRune(curr, prev protocol.TrailSegment) TrailRune {
	switch prev.Direction {
	case protocol.D_UP:
		switch curr.Direction {
		case protocol.D_UP:
			return VERTICAL
		case protocol.D_RIGHT:
			return UP_RIGHT
		case protocol.D_LEFT:
			return UP_LEFT
		}
	case protocol.D_DOWN:
		switch curr.Direction {
		case protocol.D_DOWN:
			return VERTICAL
		case protocol.D_LEFT:
			return DOWN_LEFT
		case protocol.D_RIGHT:
			return DOWN_RIGHT
		}
	case protocol.D_LEFT:
		switch curr.Direction {
		case protocol.D_LEFT:
			return HORIZONTAL
		case protocol.D_UP:
			return LEFT_UP
		case protocol.D_DOWN:
			return LEFT_DOWN
		}
	case protocol.D_RIGHT:
		switch curr.Direction {
		case protocol.D_RIGHT:
			return HORIZONTAL
		case protocol.D_DOWN:
			return RIGHT_DOWN
		case protocol.D_UP:
			return RIGHT_UP
		}
	}

	return TrailRune('0')
}

func getDirectionalPlayerRune(dir protocol.Direction) rune {
	switch dir {
	case protocol.D_UP:
		return '▲'
	case protocol.D_DOWN:
		return '▼'
	case protocol.D_LEFT:
		return '◄'
	case protocol.D_RIGHT:
		return '►'
	}

	return '1'
}

func init() {
	board := [][]rune{}
	for range 24 {
		row := []rune{}
		for range 24 {
			row = append(row, ' ')
		}
		board = append(board, row)
	}
	clearedBoard = board
}

func CreateConnection() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}

	return c
}

func Tui() {
	conn := CreateConnection()
	rm := NewRootModel(conn)
	p := tea.NewProgram(rm)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there has been an error: %v", err)
		os.Exit(1)
	}
}

func NewRootModel(conn *websocket.Conn) *RootModel {
	return &RootModel{
		Conn:     conn,
		GameComp: NewGameComponent(conn),
		Messages: make(chan protocol.GameState),
	}
}

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rm.width = msg.Width
		rm.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "w":
			rm.Conn.WriteJSON(createPlayerInput(protocol.D_UP, "player_1"))
		case "s":
			rm.Conn.WriteJSON(createPlayerInput(protocol.D_DOWN, "player_1"))
		case "a":
			rm.Conn.WriteJSON(createPlayerInput(protocol.D_LEFT, "player_1"))
		case "d":
			rm.Conn.WriteJSON(createPlayerInput(protocol.D_RIGHT, "player_1"))

		case "ctrl+c":
			return rm, tea.Quit
		}
	}
	*rm.GameComp, cmd = rm.GameComp.Update(msg)

	return rm, cmd
}

type GameStateMsg struct {
	Message protocol.GameState
}

func (rm *RootModel) View() string {
	return rm.GameComp.View()
}

func (rm *RootModel) Init() tea.Cmd {
	return tea.Batch(
		pollGame(rm.Conn),
	)
}

func pollGame(c *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var msg protocol.GameState
		err := c.ReadJSON(&msg)
		if err != nil {
			panic(err)
		}

		return GameStateMsg{Message: msg}
	}
}

func copyBoard(og [][]rune) [][]rune {
	newBoard := [][]rune{}
	for i := range og {
		newBoard = append(newBoard, []rune{})
		for _, char := range og[i] {
			newBoard[i] = append(newBoard[i], char)
		}
	}
	return newBoard
}
