package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	g := NewTicTacToe("player1", "player2")

	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		player1 := query.Get("player1")
		player2 := query.Get("player2")
		g = NewTicTacToe(player1, player2)
		w.Write([]byte(g.PrintBoard()))
	})
	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		player := query.Get("player")
		location := query.Get("location")
		if location != "" {
			if err := g.Mark(player, location); err != nil {
				w.WriteHeader(400)
				w.Write([]byte(err.Error()))
				return
			}
		}
		g.IsOver()
		board := g.PrintBoard()
		w.WriteHeader(200)
		w.Write([]byte(board))
	})

	log.Println("Starting Server...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type mark struct {
	player   int64
	location string
	turn     int
	time     time.Time
}

var locations = []string{
	"0,0", "0,1", "0,2",
	"1,0", "1,1", "1,2",
	"2,0", "2,1", "2,2",
}

var winConditions = [][]string{
	// rows
	{"0,0", "0,1", "0,2"},
	{"1,0", "1,1", "1,2"},
	{"2,0", "2,1", "2,2"},

	// cols
	{"0,0", "1,0", "2,0"},
	{"0,1", "1,1", "2,1"},
	{"0,2", "1,2", "2,2"},

	// cross
	{"0,0", "1,1", "2,2"},
	{"0,2", "1,1", "2,0"},
}

type Game struct {
	turn    int
	winner  int64
	rw      sync.RWMutex
	players map[int64]string
	board   map[string]mark
	ledger  []mark
}

func NewTicTacToe(player1, player2 string) *Game {
	return &Game{
		players: map[int64]string{
			1: player1,
			2: player2,
		},
		board:  make(map[string]mark),
		ledger: make([]mark, 0),
	}
}

func (g *Game) Mark(playerName string, location string) error {
	g.rw.Lock()
	defer g.rw.Unlock()
	playerID := int64(0)
	for id, name := range g.players {
		if name == playerName {
			playerID = id
		}
	}
	if playerID < 1 {
		return fmt.Errorf("invalid player name: %s", playerName)
	}
	if len(g.ledger) > 0 {
		if playerID == g.ledger[len(g.ledger)-1].player {
			return fmt.Errorf("not this player's turn")
		}
	}
	if err := g.validateMark(location); err != nil {
		return err
	}
	g.turn++
	m := mark{
		player:   playerID,
		location: location,
		turn:     g.turn,
		time:     time.Now(),
	}
	g.board[location] = m
	g.ledger = append(g.ledger, m)
	return nil
}

func (g *Game) PrintBoard() string {
	g.rw.RLock()
	defer g.rw.RUnlock()
	out := fmt.Sprintf("Players - %s [%d] vs %s [%d]\n", g.players[1], 1, g.players[2], 2)
	for idx, loc := range locations {
		p := int64(0)
		if m, ok := g.board[loc]; ok {
			p = m.player
		}
		out += fmt.Sprintf("[ %d ]", p)
		if (idx+1)%3 == 0 {
			out += fmt.Sprintln("")
		}
	}
	if g.winner > 0 {
		out += fmt.Sprintf("Winner is Player: %s [%d]", g.players[g.winner], g.winner)
	}
	return out
}

func (g *Game) IsOver() (bool, int64) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, condition := range winConditions {
		if done, winner := g.checkWinCondition(condition...); done {
			g.winner = winner
			return done, winner
		}
	}
	return false, 0
}

func (g *Game) validateMark(location string) error {
	if g.winner > 0 {
		return fmt.Errorf("game already over")
	}
	if m, ok := g.board[location]; ok {
		return fmt.Errorf("location: %s already taken player: %d", location, m.player)
	}

	valid := false
	for _, loc := range locations {
		if loc == location {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("location, invalid")
	}
	return nil
}

func (g *Game) checkWinCondition(loc ...string) (win bool, player int64) {
	if len(loc) != 3 {
		return false, 0
	}
	win = true
	for _, l := range loc {
		if m, ok := g.board[l]; !ok {
			return false, 0
		} else {
			if player == 0 {
				player = m.player

			}
			if m.player != player {
				win = false
				player = 0
			}
		}
	}
	return
}
