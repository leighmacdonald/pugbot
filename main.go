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
// https://discordapp.com/oauth2/authorize?&client_id=283807542751264768&scope=bot&permissions=535850256
import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"pugbot/command"
	"pugbot/config"
	"pugbot/msg"
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
		// TODO use map
		switch payload.Cmd {
		case command.CHelp:
			err = command.CmdHelp(s, m, payload)
		case command.CJoin:
			err = command.CmdJoin(s, m, payload)
		case command.CLeave:
			err = command.CmdLeave(s, m, payload)
		case command.CPlayers:
			err = command.CmdPlayers(s, m, payload)
		case command.CReset:
			err = command.CmdReset(s, m, payload)
		case command.CLimit:
			err = command.CmdLimit(s, m, payload)
		case command.CTeams:
			err = command.CmdTeams(s, m, payload)
		case command.CStart:
			err = command.CmdStart(s, m, payload)
		case command.CStop:
			err = command.CmdStop(s, m, payload)
		}
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error()}).Error("Failed to handle message")
			msg.SendErr(s, m, msg.Err(err.Error()))
		}
	}
}

func main() {
	cfg := config.GetConfig()
	SetupLogger(cfg.GetString(config.CfgLogLevel))
	discord, err := discordgo.New("Bot " + cfg.GetString(config.CfgAuthToken))
	if err != nil {
		log.Error(err)
	}
	discord.AddHandler(onMessage)

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
