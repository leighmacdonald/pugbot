package msg

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"unicode"
)

func SendErr(s *discordgo.Session, m *discordgo.MessageCreate, error string) {
	s.ChannelMessageSend(m.ChannelID, error)
}

func Err(msg string) string {
	return fmt.Sprintf("```diff\n- %s```", UcFirst(msg))
}

func Success(msg string) string {
	return fmt.Sprintf("```diff\n+ %s\n```", UcFirst(msg))
}

func Warn(msg string) string {
	return fmt.Sprintf("```fix\n> %s```", UcFirst(msg))
}

func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
