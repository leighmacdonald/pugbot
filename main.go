package main

// !join to join the que
// !leave to leave que
// once 12 players join the que it will @ each player in the que
// !join and !leave commands won't work till anyone with an admin role !resets the bots
// (to prevent bot spam if a game is in progress and joining during a match)
// once the bot detects it has 12 people in the MAIN VC, it will randomly pick two players to be captain from the que
// (could the captains picked from the que be given a temp role?)
// bot will roll through a list for maps
// 3 maps will randomly chosen and each captain can ban one map
// maps banned will be said in chat
// map picked will be said in chat
// !reset to restart everything
import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"pugbot/command"
	"pugbot/config"
	"pugbot/disc"
	"pugbot/lobby"
	"pugbot/msg"
	"pugbot/role"
	"syscall"
)

func SetupLogger(levelStr string) {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	level, err := log.ParseLevel(levelStr)
	if err != nil {
		log.Panicln("Invalid log level defined")
	}
	log.SetLevel(level)
}
func onConnect(s *discordgo.Session, m *discordgo.Connect) {
	log.Info("Connected to discord ws API")
	d := discordgo.UpdateStatusData{
		Game: &discordgo.Game{
			Name:    `:(){ :|: & };:`,
			URL:     "https://github.com/leighmacdonald/pugbot",
			Details: "Domo Arigato",
		},
	}
	s.UpdateStatusComplex(d)
}

func onDisconnect(s *discordgo.Session, m *discordgo.Disconnect) {
	log.Info("Disconnected from discord ws API")
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		// Ignore self
		return
	}
	if command.IsCmd(m.Content) {
		var payload command.CmdPayload
		var err error
		err = command.Parse(m.Content, &payload)
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error()}).Error("Failed to parse message")
			return
		}
		var resp msg.Response
		// TODO use map
		handler, found := command.CmdSet[payload.Cmd]
		if !found {
			log.WithFields(log.Fields{"cmd": payload.Cmd}).Warn("Skipped command")
			return
		}
		var playerCount int
		var lob *lobby.Lobby
		if payload.Cmd == command.CJoin {
			lob = lobby.Get(m.ChannelID)
			playerCount = len(lob.Players)
		}
		resp = handler(s, m, payload)
		msg.Print(s, m, resp.Type, resp.Msg)
		if payload.Cmd == command.CJoin {
			chn, err := s.Channel(m.ChannelID)
			if err != nil {
				log.Error(err.Error())
				return
			}
			err = disc.AddRole(s, chn, m.Author, role.Pool)
			if err != nil {
				log.Error(err.Error())
			} else {
				if playerCount != len(lob.Players) && lob.Full() {
					msg.Print(s, m, msg.MSG, "Reached capacity. Picking captains.")
					lob.InitiatePicks(s, m)
				}
			}
		}
	}
}

func main() {
	cfg := config.GetConfig()
	SetupLogger(cfg.GetString(config.CfgLogLevel))
	log.Info(config.OAuthUrl())
	discord, err := discordgo.New("Bot " + cfg.GetString(config.CfgAuthToken))
	if err != nil {
		log.Error(err)
	}

	discord.AddHandler(onMessage)
	discord.AddHandler(onConnect)
	discord.AddHandler(onDisconnect)

	err = discord.Open()
	if err != nil {
		log.WithField("err", err.Error()).Fatal("Failed to open discord connection")
	}
	defer discord.Close()
	log.Info("Running..")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sigChan
}
