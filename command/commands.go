package command

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"pugbot/lobby"
	"pugbot/msg"
	"strconv"
	"strings"
)

type Command int

const (
	CmdPrefix = "!"

	CHelp    Command = iota
	CJoin    Command = iota
	CLeave   Command = iota
	CPick    Command = iota
	CBan     Command = iota
	CReset   Command = iota
	CPlayers Command = iota
	CUnknown Command = iota
	CTeams   Command = iota
	CLimit   Command = iota
	CStart   Command = iota
	CStop    Command = iota
)

type CmdPayload struct {
	Cmd Command
	Msg string
}

func Parse(msg string, payload *CmdPayload) error {
	var err error
	pieces := strings.SplitN(msg[1:], " ", 2)
	if len(pieces) > 1 {
		payload.Msg = pieces[1]
	}
	switch strings.ToLower(pieces[0]) {
	case "help":
		payload.Cmd = CHelp
	case "join":
		payload.Cmd = CJoin
	case "leave":
		payload.Cmd = CLeave
	case "ban":
		payload.Cmd = CBan
	case "pick":
		payload.Cmd = CPick
	case "players":
		payload.Cmd = CPlayers
	case "reset":
		payload.Cmd = CReset
	case "limit":
		payload.Cmd = CLimit
	case "teams":
		payload.Cmd = CTeams
	case "start":
		payload.Cmd = CStart
	case "stop":
		payload.Cmd = CStop
	default:
		payload.Cmd = CUnknown
	}
	return err
}

func IsCmd(msg string) bool {
	return len(msg) > 0 && strings.HasPrefix(msg, CmdPrefix)
}

func CmdLeave(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	l := lobby.Get(m.ChannelID)
	err := l.Leave(m)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageSend(m.ChannelID, msg.Err(fmt.Sprintf("%s left lobby", m.Author.Username)))
	return err
}

func CmdJoin(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	l := lobby.Get(m.ChannelID)
	if !l.Started {
		return errors.New("no lobby started")
	}
	if l.Full() {
		return errors.New("lobby currently full")
	}
	err := l.Join(m)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageSend(m.ChannelID, msg.Success(fmt.Sprintf("%s joined lobby", m.Author.Username)))
	if l.Full() {
		_, err = s.ChannelMessageSend(m.ChannelID, msg.Warn(fmt.Sprintf("Player limit reached (%d)", l.Limit)))
		return err
	}
	return err
}

func CmdPlayers(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	l := lobby.Get(m.ChannelID)
	players := l.PlayerList()
	var resp string
	if len(players) == 0 {
		resp = "no players joined"
	} else {
		resp = strings.Join(players, ", ")
	}
	_, err := s.ChannelMessageSend(m.ChannelID, msg.Warn(fmt.Sprintf("`Current lobby players: %s", resp)))
	return err
}

func CmdReset(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	l := lobby.Get(m.ChannelID)
	err := l.Reset()
	if err != nil {
		msg.SendErr(s, m, err.Error())
	} else {
		s.ChannelMessageSend(m.ChannelID, "```diff\n+ Lobby reset successfully")
	}
	return nil
}

func CmdHelp(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	help := `!help !join !leave !pick !ban !reset`
	s.ChannelMessageSend(m.ChannelID, msg.Warn(help))
	return nil
}

func CmdLimit(s *discordgo.Session, m *discordgo.MessageCreate, payload CmdPayload) error {
	l := lobby.Get(m.ChannelID)

	if payload.Msg == "" {
		s.ChannelMessageSend(m.ChannelID, msg.Success(fmt.Sprintf(
			"Current player limits: %d", l.Limit)))
		return nil
	}

	oldLimit := l.Limit
	newLimit, err := strconv.Atoi(payload.Msg)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		return errors.New("failed to parse limit value")
	}
	err = l.SetLimit(newLimit)
	if err != nil {
		return err
	}
	s.ChannelMessageSend(m.ChannelID, msg.Success(fmt.Sprintf(
		"Updated max players from %d -> %d", oldLimit, newLimit)))
	return nil
}

func CmdTeams(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	v := "```" +
		"Red Team   | Blue Team" +
		"---------- | ----------" +
		"123       | 345```"
	s.ChannelMessageSend(m.ChannelID, v)
	return nil
}

func CmdStart(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	l := lobby.Get(m.ChannelID)
	if l.Started {
		return errors.New("lobby already started")
	}
	l.Start()
	s.ChannelMessageSend(m.ChannelID, msg.Success("Lobby started! Anyone who wants in can !join now!"))
	return nil
}

func CmdStop(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) error {
	l := lobby.Get(m.ChannelID)
	if !l.Started {
		return errors.New("lobby not started")
	}
	l.Stop()
	s.ChannelMessageSend(m.ChannelID, msg.Success("Lobby successfully stopped. Thanks for playing"))
	return nil
}
