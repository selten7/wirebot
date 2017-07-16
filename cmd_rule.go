package wirebot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (wb *Wirebot) cmdRule(ds *discordgo.Session, m *discordgo.MessageCreate, trailing string) error {
	var (
		err error

		rules []string
		parts []string
	)

	rawGuildID, ok := wb.channelGuild(m.ChannelID)
	if !ok {
		return fmt.Errorf("missing guild for channel %#v", m.ChannelID)
	}

	guildID := GuildID(rawGuildID)

	rules = wb.State.GuildRules[guildID]

	for _, p := range strings.Split(trailing, " ") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}

	if len(parts) == 0 {
		return nil
	}

	action := parts[0]
	parts = parts[1:]

	var (
		n int
		o int
	)

	switch action {
	case "list":
		if len(parts) != 0 {
			return errors.New("expecting no arguments")
		}

		lines := make([]string, len(rules))
		for i, rule := range rules {
			lines[i] = fmt.Sprintf("%d. %s", i+1, rule)
		}

		msg := strings.Join(lines, "\n")

		_, err = ds.ChannelMessageSend(m.ChannelID, msg)

		return err

	case "get":
		fallthrough
	case "show":
		if len(parts) != 1 {
			return errors.New("expecting exactly one argument")
		}

		n, err = strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		i := n - 1
		if i < 0 || i >= len(rules) {
			return fmt.Errorf("rule %d not found", n)
		}

		msg := fmt.Sprintf("Rule #%d: %s", n, rules[i])

		_, err = ds.ChannelMessageSend(m.ChannelID, msg)

		return err

	case "replace":
		fallthrough
	case "set":
		if len(parts) < 2 {
			return errors.New("expecting 2 arguments")
		}

		n, err = strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		if n < 1 {
			return errors.New("number must be > 0")
		}

		i := n - 1
		if i > len(rules) {
			return fmt.Errorf("number %d is too high", n)
		}

		text := strings.Join(parts[1:], " ")

		if i == len(rules) {
			rules = append(rules, text)

			msg := fmt.Sprintf("Added rule #%d: %s", n, text)

			_, err = ds.ChannelMessageSend(m.ChannelID, msg)
		} else {
			oldRule := rules[i]
			rules[i] = text

			msg := fmt.Sprintf("Replaced rule #%d.\n\nOld: %s\nNew: %s", n, oldRule, text)

			_, err = ds.ChannelMessageSend(m.ChannelID, msg)
		}

		return err

	case "del":
		fallthrough
	case "delete":
		if len(parts) != 1 {
			return errors.New("expecting exactly 1 argument")
		}

		n, err = strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		i := n - 1
		if i < 0 || i >= len(rules) {
			return fmt.Errorf("rule %d not found", n)
		}

		if i == 0 {
			rules = rules[1:]
		} else if i+1 == len(rules) {
			rules = rules[:len(rules)-1]
		} else {
			newRules := make([]string, 0, len(rules)-1)

			newRules = append(newRules, rules[:i]...)
			newRules = append(newRules, rules[i+1:]...)

			rules = newRules
		}

	case "mv":
		fallthrough
	case "move":
		if len(parts) != 2 {
			return errors.New("expecting exactly 2 arguments")
		}

		n, err = strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		o, err = strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		i := n - 1
		j := o - 1

		if i < 0 || i >= len(rules) {
			return fmt.Errorf("invalid \"from\" position: %d", n)
		}
		if j < 0 || j >= len(rules) {
			return fmt.Errorf("invalid \"to\" position: %d", o)
		}

		var msg string

		if i == j {
			msg = fmt.Sprintf("Nothing changed.")
		} else {
			msg = fmt.Sprintf("Moved rule #%d to position #%d", n, o)

			newRules := make([]string, 0, len(rules))

			if i > j {
				newRules = append(newRules, rules[:j]...)
				newRules = append(newRules, rules[i])
				newRules = append(newRules, rules[j:i]...)
				newRules = append(newRules, rules[i+1:]...)
			} else {
				newRules = append(newRules, rules[:i]...)
				newRules = append(newRules, rules[j])
				newRules = append(newRules, rules[i:j]...)
				newRules = append(newRules, rules[j+1:]...)
			}
		}

		_, err = ds.ChannelMessageSend(m.ChannelID, msg)

		return err

	case "append":
		fallthrough
	case "add":
		if len(parts) == 0 {
			return errors.New("expecting one argument")
		}

		text := strings.Join(parts, " ")

		rules = append(rules, text)

		msg := fmt.Sprintf("Added rule #%d: %s", len(rules), text)

		_, err = ds.ChannelMessageSend(m.ChannelID, msg)

		return err

	default:
		return nil
	}

	wb.State.GuildRules[guildID] = rules

	return wb.State.Sync()
}
