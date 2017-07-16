package wirebot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrMissingMuteRole = errors.New("missing role for muting")
)

var defaultMutedRoleNames = []string{
	"mute",
	"muted",
	"silence",
	"silenced",
}

func isMutedRole(roleName string) bool {
	for _, r := range defaultMutedRoleNames {
		if strings.ToLower(r) == strings.ToLower(roleName) {
			return true
		}
	}

	return false
}

func (wb *Wirebot) cmdMute(ds *discordgo.Session, m *discordgo.MessageCreate, trailing string) error {
	var (
		err error
		i   int
	)

	// Remove mentions from message (we'll get them from m.Mentions).
	parts := strings.Split(trailing, " ")
	for i = 0; i < len(parts); i++ {
		part := parts[i]
		if !strings.HasPrefix(part, "<@") || !strings.HasSuffix(part, ">") {
			continue
		}

		part = strings.TrimPrefix(part, "<@")
		part = strings.TrimSuffix(part, ">")

		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		_, err = strconv.ParseInt(part, 10, 64)
		if err != nil {
			break
		}
	}

	parts = parts[i:]

	var (
		d           time.Duration
		dStr        string
		brokenRules []int
		rulesStr    string
	)

	switch len(parts) {
	case 0:
		return nil
	case 1:
		dStr = parts[0]
	case 2:
		dStr = parts[0]
		rulesStr = parts[1]
	default:
		return fmt.Errorf("unexpected argument %#v", parts[2])
	}

	d, err = time.ParseDuration(dStr)
	if err != nil {
		return err
	}

	// Get list of broken rules, if applicable.
	if rulesStr != "" {
		for _, nStr := range strings.Split(rulesStr, ",") {
			if strings.HasPrefix(nStr, "#") {
				nStr = strings.TrimPrefix(nStr, "#")
			}

			var n int

			n, err = strconv.Atoi(nStr)
			if err != nil {
				return err
			}

			brokenRules = append(brokenRules, n)
		}
	}

	guildID, ok := wb.channelGuild(m.ChannelID)
	if !ok {
		return fmt.Errorf("missing guild for channel %#v", m.ChannelID)
	}

	mutedMentions := make([]string, len(m.Mentions))
	for i, mention := range m.Mentions {
		mutedMentions[i] = mention.Mention()
	}

	// Mute the mentioned users.
	for _, userID := range mutedMentions {
		err = wb.muteUser(guildID, userID, d)
		if err != nil {
			return err
		}
	}

	if len(brokenRules) == 0 {
		response := fmt.Sprintf("**Muted:** %s", strings.Join(mutedMentions, " "))

		_, err = ds.ChannelMessageSend(m.ChannelID, response)

		return err
	}

	var (
		rulesSlice []string
		rulesMap   map[int]string // Map rule numbers to their content.
	)

	if wb.State != nil && wb.State.GuildRules != nil {
		rulesSlice = wb.State.GuildRules[GuildID(guildID)]
	}

	var key int
	for _, ruleNumber := range brokenRules {
		i = ruleNumber - 1
		if i >= len(rulesSlice) {
			return fmt.Errorf("rule #%d not found", ruleNumber)
		}
		if rulesMap == nil {
			rulesMap = make(map[int]string)
		}

		rulesMap[ruleNumber] = rulesSlice[i]
	}

	rulesList := make([]string, len(brokenRules))
	for i, key = range brokenRules {
		rulesList[i] = fmt.Sprintf("#%d. %s", key, rulesMap[key])
	}

	response := fmt.Sprintf(
		"**Muted:** %s\n\nRules broken: %s",
		strings.Join(mutedMentions, " "),
		strings.Join(rulesList, "\n"),
	)

	_, err = ds.ChannelMessageSend(m.ChannelID, response)

	return nil
}

func (wb *Wirebot) muteUser(guildID, userID string, duration time.Duration) error {
	guildRoles, err := wb.session.GuildRoles(guildID)
	if err != nil {
		return err
	}

	var roleID string

	for _, gr := range guildRoles {
		if isMutedRole(gr.Name) {
			roleID = gr.ID

			break
		}
	}

	if roleID == "" {
		return ErrMissingMuteRole
	}

	err = wb.session.GuildMemberRoleAdd(guildID, userID, roleID)
	if err != nil {
		return err
	}

	mutedUntil := time.Now().Add(duration)

	if wb.State.MutedUsers == nil {
		wb.State.MutedUsers = make(map[GuildID][]*UserMute)
	}
	for _, um := range wb.State.MutedUsers[GuildID(guildID)] {
		if um.UserID == roleID {
			um.MutedUntil = mutedUntil

			return nil
		}
	}

	wb.State.MutedUsers[GuildID(guildID)] = append(wb.State.MutedUsers[GuildID(guildID)], &UserMute{
		UserID:     userID,
		MutedUntil: mutedUntil,
	})

	wb.logger().Printf("muted user %s on guild %s", userID, guildID)

	return nil
}

func (wb *Wirebot) unmuteUser(guildID, userID string) error {
	guildRoles, err := wb.session.GuildRoles(guildID)
	if err != nil {
		return err
	}

	var roleID string

	for _, gr := range guildRoles {
		if isMutedRole(gr.Name) {
			roleID = gr.ID

			break
		}
	}

	if roleID == "" {
		return ErrMissingMuteRole
	}

	err = wb.session.GuildMemberRoleRemove(guildID, userID, roleID)
	if err != nil {
		return err
	}

	if wb.State.MutedUsers == nil {
		return nil
	}

	idx := -1
	for i, um := range wb.State.MutedUsers[GuildID(guildID)] {
		if um.UserID == roleID {
			idx = i

			break
		}
	}

	if idx != -1 {
		wb.State.MutedUsers[GuildID(guildID)] = append(
			wb.State.MutedUsers[GuildID(guildID)][:idx],
			wb.State.MutedUsers[GuildID(guildID)][idx+1:]...,
		)
	}

	return nil
}
