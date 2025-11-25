package client

import (
	"encoding/json"
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

	Conn             *websocket.Conn
	GameComp         *GameComponent
	Messages         chan (protocol.GameState)
	PlayerAssignment string
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
	Players   map[string]protocol.PlayerState
	Board     [][]rune
	Tick      int
	Conn      *websocket.Conn
	Countdown int
}

func NewGameComponent(conn *websocket.Conn) *GameComponent {
	return &GameComponent{
		Players:   make(map[string]protocol.PlayerState),
		Board:     clearedBoard,
		Conn:      conn,
		Countdown: 0,
	}
}

func (g *GameComponent) View() string {
	boardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))
	s := fmt.Sprintf("tick: %d\n", g.Tick)
	s += fmt.Sprintf("player1: %d, %d, %s\n", g.Players["player_1"].Position.X, g.Players["player_1"].Position.Y, g.Players["player_1"].Status)
	s += fmt.Sprintf("score:\n\t player1: %d\n\t player2: %d\n", g.Players["player_1"].Points, g.Players["player_2"].Points)
	if g.Countdown > 0 {
		s += fmt.Sprintf("Game starting in: %d seconds\n", g.Countdown)
	} else {
		s += "\n"
	}
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
	g.Countdown = msg.Message.Countdown
	board := copyBoard(clearedBoard)
	for _, player := range g.Players {
		px := player.Position.X
		py := player.Position.Y
		if player.Status == "DEAD" {
			if board[py][px] == 'x' {
				board[py][px] = 'җ'
			} else {
				board[py][px] = 'x'
			}
		} else {
			board[py][px] = getDirectionalPlayerRune(player.Direction)
		}
		for i, trailSeg := range player.Trail {
			if i == 0 {
				continue
			}
			tx := trailSeg.Coordinate.X
			ty := trailSeg.Coordinate.Y
			if board[ty][tx] == 'x' {
				continue
			}
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
		for range 48 {
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
	case PlayerAssignmentMsg:
		rm.PlayerAssignment = msg.Message.PlayerID
		return rm, pollGame(rm.GameComp.Conn)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			rm.Conn.WriteJSON(createGameCommand("start"))
		case "w", "up":
			rm.Conn.WriteJSON(createPlayerMessage(protocol.D_UP, rm.PlayerAssignment))
		case "s", "down":
			rm.Conn.WriteJSON(createPlayerMessage(protocol.D_DOWN, rm.PlayerAssignment))
		case "a", "left":
			rm.Conn.WriteJSON(createPlayerMessage(protocol.D_LEFT, rm.PlayerAssignment))
		case "d", "right":
			rm.Conn.WriteJSON(createPlayerMessage(protocol.D_RIGHT, rm.PlayerAssignment))

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
type PlayerAssignmentMsg struct {
	Message protocol.PlayerAssignment
}

// TODO: We need to show rootModel information above game information. OR to the left or whatever
func (rm *RootModel) View() string {
	left := rm.RenderLeft()
	game := rm.GameComp.View()
	return lipgloss.JoinHorizontal(lipgloss.Top, game, left)
}

func (rm *RootModel) RenderGame() string {
	content := rm.GameComp.View()
	style := lipgloss.NewStyle().
		Width(rm.width / 2)
	return style.Render(content)
}

func (rm RootModel) RenderLeft() string {
	content := fmt.Sprintf("PlayerID: %s", rm.PlayerAssignment)
	headerStyle := lipgloss.NewStyle().
		Width(rm.width / 2).
		Height(rm.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))
	return headerStyle.Render(content)
}

func (rm *RootModel) Init() tea.Cmd {
	return tea.Batch(
		pollGame(rm.Conn),
	)
}

func pollGame(c *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var msg protocol.Message
		err := c.ReadJSON(&msg)
		if err != nil {
			panic(err)
		}
		// return GameStateMsg{Message: msg}
		return processMessage(msg)
	}
}

func processMessage(msg protocol.Message) any {
	switch msg.Type {
	case "GameState":
		var body protocol.GameState
		json.Unmarshal(msg.Body, &body)
		return GameStateMsg{Message: body}
	case "PlayerAssignment":
		var body protocol.PlayerAssignment
		json.Unmarshal(msg.Body, &body)
		return PlayerAssignmentMsg{Message: body}
	}

	return nil
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
