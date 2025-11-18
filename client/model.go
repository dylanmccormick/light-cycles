package client

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanmccormick/light-cycles/protocol"
	"github.com/gorilla/websocket"
)

type RootModel struct {
	width  int
	height int

	GameComp *GameComponent
	Conn     *websocket.Conn
	Messages chan (protocol.GameState)
}

var clearedBoard [][]rune

type GameComponent struct {
	Players map[string]protocol.PlayerState
	Board   [][]rune
	Tick    int
}

func NewGameComponent() *GameComponent {
	return &GameComponent{
		Players: make(map[string]protocol.PlayerState),
		Board:   clearedBoard,
	}
}

func (g *GameComponent) drawBoard(msg GameStateMsg) {
	g.Players = msg.Message.Players
	g.Tick = msg.Message.Tick
	board := clearedBoard
	for _, player := range g.Players {
		px := player.Position.X
		py := player.Position.Y
		board[px][py] = '1'
		for _, co := range player.Trail {
			tx := co.X
			ty := co.Y
			board[tx][ty] = '-'
		}
	}
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

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rm.width = msg.Width
		rm.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return rm, tea.Quit
		}
	case GameStateMsg:
		rm.GameComp.drawBoard(msg)
	}

	return rm, cmd
}

type GameStateMsg struct {
	Message protocol.GameState
}

func (rm *RootModel) View() string {
	return ""
}

func (rm *RootModel) Init() tea.Cmd {
	return tea.Batch{
		pollGame(rm.Conn),
	}
}

func pollGame(c *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		for {
			var msg protocol.GameState
			err := c.ReadJSON(&msg)
			if err != nil {
				panic(err)
			}

			return GameStateMsg{Message: msg}
		}
	}
}
