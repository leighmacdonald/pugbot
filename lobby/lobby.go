package lobby

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"sort"
	"sync"
)

const (
	limitDefault = 12
)

var (
	lobbies Lobbies
)

type Lobbies struct {
	lobbies map[string]*Lobby
	lock    *sync.Mutex
}

type Player struct {
	ID      string
	Name    string
	Captain bool
}

type Lobby struct {
	Limit    int
	Started  bool
	TeamRed  []*Player
	TeamBlue []*Player
	Players  []*Player
}

func (l *Lobby) SetLimit(limit int) error {
	if limit <= 0 || limit > 18 {
		return errors.New("invalid limit must be between 2-18")
	} else if limit%2 != 0 {
		return errors.New("limit value must be even number")
	} else {
		l.Limit = limit
		return nil
	}
}
func (l *Lobby) Full() bool {
	return len(l.Players) >= l.Limit
}

func (l *Lobby) PlayerList() []string {
	var players []string
	for _, p := range l.Players {
		players = append(players, p.Name)
	}
	sort.Strings(players)
	return players
}

func (l *Lobby) Join(m *discordgo.MessageCreate) error {
	for _, p := range l.Players {
		if p.ID == m.Author.ID {
			return errors.New("already joined lobby")
		}
	}
	l.Players = append(l.Players, &Player{
		m.Author.ID,
		m.Author.Username,
		false,
	})
	return nil
}

func NewLobbies() Lobbies {
	return Lobbies{
		lobbies: make(map[string]*Lobby),
		lock:    new(sync.Mutex),
	}
}

func (l *Lobbies) Get(channelId string) *Lobby {
	l.lock.Lock()
	defer l.lock.Unlock()
	lob, found := l.lobbies[channelId]
	if !found {
		l.lobbies[channelId] = &Lobby{Limit: limitDefault, Started: false}
		lob = l.lobbies[channelId]
	}
	return lob
}

func (l *Lobby) Start() {
	l.Started = true
}
func (l *Lobby) Stop() {
	l.Started = false
}

func (l *Lobby) Leave(m *discordgo.MessageCreate) error {
	oldCnt := len(l.Players)
	l.Players = removePlayer(l.Players, m.Author.ID)
	if oldCnt == len(l.Players) {
		return errors.New("player doesn't exist")
	}
	return nil
}

func (l *Lobby) Reset() error {
	l.TeamBlue = nil
	l.TeamRed = nil
	l.Players = nil
	return nil
}

func removePlayer(players []*Player, playerId string) []*Player {
	var newPlayers []*Player
	for _, p := range players {
		if p.ID != playerId {
			newPlayers = append(newPlayers, p)
		}
	}
	return newPlayers
}

func Get(channelId string) *Lobby {
	return lobbies.Get(channelId)
}

func init() {
	lobbies = NewLobbies()
}
