package disc

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"pugbot/role"
	"strings"
)

func FindRole(s *discordgo.Session, guildID string, name role.Role) *discordgo.Role {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	for _, r := range roles {
		if strings.ToLower(r.Name) == strings.ToLower(string(name)) {
			return r
		}
	}
	return nil
}

func AddRole(s *discordgo.Session, c *discordgo.Channel, u *discordgo.User, name role.Role) error {
	r := FindRole(s, c.GuildID, name)
	if r == nil {
		return errors.New("failed to find role")
	}
	err := s.GuildMemberRoleAdd(c.GuildID, u.ID, r.ID)
	log.WithFields(log.Fields{
		"role": name,
		"user": u.Username,
	}).Debug("Added role to user")
	return err
}
