package msg

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"unicode"
)

type ResponseType int

const (
	MSG  ResponseType = iota
	WARN ResponseType = iota
	ERR  ResponseType = iota
)

type Response struct {
	Msg  string
	Type ResponseType
}

func ToString(typ ResponseType, msg string) string {
	switch typ {
	case ERR:
		return Err(msg)
	case WARN:
		return Warn(msg)
	case MSG:
		return Success(msg)
	default:
		return msg
	}
}

func Print(s *discordgo.Session, m *discordgo.MessageCreate, t ResponseType, msg string) {
	s.ChannelMessageSend(m.ChannelID, ToString(t, msg))
}

func (r *Response) Error() string {
	return r.Msg
}

func NewError(msg string) Response {
	return Response{
		Msg:  msg,
		Type: ERR,
	}
}

func NewWarn(msg string) Response {
	return Response{
		Msg:  msg,
		Type: WARN,
	}
}

func NewMsg(msg string) Response {
	return Response{
		Msg:  msg,
		Type: MSG,
	}
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
