package command

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"pugbot/disc"
	"pugbot/lobby"
	"pugbot/msg"
	"pugbot/role"
	"strconv"
	"strings"
)

type Command string

const (
	CmdPrefix = "!"

	CUnknown Command = "unknown"
	CHelp    Command = "help"
	CJoin    Command = "join"
	CLeave   Command = "leave"
	CPick    Command = "pick"
	CBan     Command = "ban"
	CReset   Command = "reset"
	CPlayers Command = "players"
	CTeams   Command = "teams"
	CLimit   Command = "limit"
	CStart   Command = "start"
	CStop    Command = "stop"
	CAssign  Command = "assign"
	CStatus  Command = "status"
	CRoles   Command = "roles"
)

func AvailableCommandList() []string {
	return []string{"help", "join", "leave", "pick", "ban", "reset", "players", "teams",
		"limit", "start", "stop", "assign", "roles", "status"}
}

type ChatCommandI func(s *discordgo.Session, m *discordgo.MessageCreate, p CmdPayload) msg.Response

var CmdSet map[Command]ChatCommandI

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
	// TODO dont be a meme.
	cmdStr := strings.ToLower(pieces[0])
	payload.Cmd = Command(cmdStr)
	return err
}

func IsCmd(msg string) bool {
	return len(msg) > 0 && strings.HasPrefix(msg, CmdPrefix)
}

func CmdLeave(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	err := l.Leave(m)
	if err != nil {
		return msg.NewError(err.Error())
	}
	return msg.NewWarn(fmt.Sprintf("%s left lobby", m.Author.Username))
}

func CmdJoin(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	if !l.Started {
		return msg.NewError("no lobby started")
	}
	if l.Full() {
		return msg.NewError("lobby currently full")
	}
	err := l.Join(m)
	if err != nil {
		return msg.NewError(err.Error())
	}
	return msg.NewMsg(fmt.Sprintf("%s joined lobby", m.Author.Username))
}

func CmdPlayers(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	players := l.PlayerList()
	var resp string
	if len(players) == 0 {
		resp = "no players joined"
	} else {
		resp = strings.Join(players, ", ")
	}
	return msg.NewWarn(fmt.Sprintf("`Current lobby players: %s", resp))
}

func CmdReset(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	err := l.Reset()
	if err != nil {
		return msg.NewError(err.Error())
	} else {
		return msg.NewMsg("Lobby reset successfully")
	}
}

func CmdHelp(_ *discordgo.Session, _ *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	help := CmdPrefix + strings.Join(AvailableCommandList(), ", "+CmdPrefix)
	return msg.NewWarn(help)
}

func CmdLimit(_ *discordgo.Session, m *discordgo.MessageCreate, payload CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	if payload.Msg == "" {
		return msg.NewMsg(fmt.Sprintf(
			"Current player limits: %d", l.Limit))
	}
	oldLimit := l.Limit
	newLimit, err := strconv.Atoi(payload.Msg)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err.Error()}).Error(err)
		return msg.NewError("failed to parse limit value")
	}
	err = l.SetLimit(newLimit)
	if err != nil {
		return msg.NewError(err.Error())
	}
	return msg.NewMsg(fmt.Sprintf(
		"Updated max players from %d -> %d", oldLimit, newLimit))
}

func CmdTeams(_ *discordgo.Session, _ *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	v := "```" +
		"Red Team   | Blue Team" +
		"---------- | ----------" +
		"123       | 345```"
	return msg.NewMsg(v)
}

func CmdStart(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	if l.Started {
		return msg.NewWarn("lobby already started")
	}
	l.Start()
	return msg.NewMsg("Lobby started! Anyone who wants in can !join now!")
}

func CmdStop(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	if !l.Started {
		return msg.NewError("lobby not started")
	}
	l.Stop()
	return msg.NewMsg("Lobby successfully stopped. Thanks for playing")
}

func CmdPick(_ *discordgo.Session, _ *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	return msg.NewError("pick not implemented")
}

func CmdBan(_ *discordgo.Session, _ *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	return msg.NewError("ban not implemented")
}

func CmdStatus(_ *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	l := lobby.Get(m.ChannelID)
	return msg.NewMsg(l.Status())
}

func CmdAssign(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	values := strings.Split(m.Content, " ")
	if len(values) != 3 {
		return msg.NewError("Malformed request")
	}
	team := values[1]
	if team == "blue" {
		team = "blu"
	}
	chn, err := s.Channel(m.ChannelID)
	if err != nil {
		return msg.NewError(err.Error())
	}
	roleRed := disc.FindRole(s, chn.GuildID, role.Red)
	roleBlu := disc.FindRole(s, chn.GuildID, role.Blu)
	var errRem error
	if team == "red" {
		err = s.GuildMemberRoleAdd(chn.GuildID, m.Mentions[0].ID, roleRed.ID)
		errRem = s.GuildMemberRoleRemove(chn.GuildID, m.Mentions[0].ID, roleBlu.ID)
	} else {
		err = s.GuildMemberRoleAdd(chn.GuildID, m.Mentions[0].ID, roleBlu.ID)
		errRem = s.GuildMemberRoleRemove(chn.GuildID, m.Mentions[0].ID, roleRed.ID)
	}
	if errRem != nil {
		logrus.WithField("err", errRem.Error()).Warn("Couldn't remove user from role")
	}
	if err != nil {
		return msg.NewError(err.Error())
	} else {
		return msg.NewMsg(fmt.Sprintf("Added %s to %s", m.Mentions[0].Username, team))
	}

	return msg.NewError("Failed to find role?")
}

func CmdRoles(s *discordgo.Session, m *discordgo.MessageCreate, _ CmdPayload) msg.Response {
	var roles []string
	chn, err := s.Channel(m.ChannelID)
	if err != nil {
		return msg.NewError(err.Error())
	}
	knownRoles, err := s.GuildRoles(chn.GuildID)
	if err != nil {
		return msg.NewError(err.Error())
	}
	for _, r := range knownRoles {
		roles = append(roles, r.Name)
	}
	return msg.NewMsg(fmt.Sprintf("Server roles: %s", strings.Join(roles, ", ")))
}

func init() {
	CmdSet = map[Command]ChatCommandI{
		CHelp:    CmdHelp,
		CJoin:    CmdJoin,
		CLeave:   CmdLeave,
		CPick:    CmdPick,
		CBan:     CmdBan,
		CReset:   CmdReset,
		CPlayers: CmdPlayers,
		CTeams:   CmdTeams,
		CLimit:   CmdLimit,
		CStart:   CmdStart,
		CStop:    CmdStop,
		CAssign:  CmdAssign,
		CRoles:   CmdRoles,
		CStatus:  CmdStatus,
	}
}
