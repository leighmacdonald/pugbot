package lobby

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"pugbot/disc"
	"pugbot/msg"
	"sort"
	"sync"
)

type Team string

const (
	RED          Team = "red"
	BLU          Team = "blu"
	limitDefault      = 12
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
	Limit       int
	Started     bool
	Teams       map[Team][]*Player
	Players     []*Player
	CurrentPick Team
}

func (l *Lobby) SelectRandPicker() Team {
	if rand.Intn(1) == 0 {
		l.CurrentPick = RED
	} else {
		l.CurrentPick = BLU
	}
	return l.CurrentPick
}

func (l *Lobby) Status() string {
	// No Lobby
	if !l.Started {
		return "No lobby active. type !start to initiate one."
	}
	// Awaiting players
	if !l.Full() {
		return fmt.Sprintf("Lobby started. Waiting on players [%d/%d]", len(l.Players), l.Limit)
	} else {
		assignedPlayers := len(l.Teams[RED]) + len(l.Teams[BLU])
		// Awaiting captain picks
		if assignedPlayers < len(l.Players) {
			return fmt.Sprintf(
				"Waiting for captains to pick. Current pick: %s Remaining picls: %d",
				l.CurrentPick,
				len(l.Players)-assignedPlayers)
		} else {
			// Awaiting map voting
			return fmt.Sprintf(
				"Teams assigned. Waiting on map votes")
		}
	}
}

func (l *Lobby) InitiatePicks(s *discordgo.Session, m *discordgo.MessageCreate) error {
	err := l.PickCaptains()
	if err != nil {
		msg.Print(s, m, msg.ERR, "Error picking captains RIP :(")
		return err
	}

	info := fmt.Sprintf("BLU Captain: %s  --  RED Captain: %s",
		l.GetCaptain(BLU).Name,
		l.GetCaptain(RED).Name)
	msg.Print(s, m, msg.MSG, info)
	chn, err := s.Channel(m.ChannelID)
	err = disc.AddRole(s, chn, m.Author, "Red Team")

	l.SelectRandPicker()

	msg.Print(s, m, msg.MSG,
		fmt.Sprintf("First pick goes to %s team (%s)",
			l.CurrentPick,
			l.GetCaptain(l.CurrentPick).Name))
	return err
}

func (l *Lobby) PickCaptains() error {
	var bluI int
	var redI int
	redI = rand.Intn(len(l.Players) - 1)
	bluI = rand.Intn(len(l.Players) - 1)
	for redI == bluI {
		bluI = rand.Intn(len(l.Players))
	}
	redC := l.Players[redI]
	redC.Captain = true
	l.Teams[RED] = append(l.Teams[RED], redC)
	bluC := l.Players[bluI]
	bluC.Captain = true
	l.Teams[BLU] = append(l.Teams[BLU], bluC)
	log.WithFields(log.Fields{
		"red_captain": redC.Name,
		"blu_captain": bluC.Name,
	}).Info("Picked captains")
	return nil
}

func (l *Lobby) GetCaptain(team Team) *Player {
	for _, player := range l.Teams[team] {
		if player.Captain {
			return player
		}
	}
	log.Warn("No captains found")
	return nil
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

func NewPlayer(id string, name string) *Player {
	return &Player{
		id,
		name,
		false,
	}
}

func (l *Lobby) Join(m *discordgo.MessageCreate) error {
	for _, p := range l.Players {
		if p.ID == m.Author.ID {
			return errors.New("already joined lobby")
		}
	}
	l.Players = append(l.Players, NewPlayer(m.Author.ID, m.Author.Username))
	return nil
}

func NewLobbies() Lobbies {
	return Lobbies{
		lobbies: make(map[string]*Lobby),
		lock:    new(sync.Mutex),
	}
}

func NewLobby() *Lobby {
	return &Lobby{
		Teams:   map[Team][]*Player{},
		Limit:   limitDefault,
		Started: false,
	}
}
func (l *Lobbies) Get(channelId string) *Lobby {
	l.lock.Lock()
	defer l.lock.Unlock()
	lob, found := l.lobbies[channelId]
	if !found {
		l.lobbies[channelId] = NewLobby()
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
	for _, p := range l.Players {
		p.Captain = false
	}
	l.Teams[RED] = nil
	l.Teams[BLU] = nil
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
