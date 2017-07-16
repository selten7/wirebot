package wirebot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// OnReady is the handler for the READY event.
func (wb *Wirebot) OnReady(ds *discordgo.Session, ready *discordgo.Ready) {
	wb.State = new(State)
	wb.guilds = ready.Guilds

	wb.logger().Print("ready")
}

// OnGuildCreate is the handler for the GUILD_CREATE event.
func (wb *Wirebot) OnGuildCreate(ds *discordgo.Session, gc *discordgo.GuildCreate) {
	g := gc.Guild

	for i, wbg := range wb.guilds {
		if g.ID == wbg.ID {
			wb.guilds[i] = g

			return
		}
	}

	wb.guilds = append(wb.guilds, g)
}

// OnGuildDelete is the handler for the GUILD_DELETE event.
func (wb *Wirebot) OnGuildDelete(ds *discordgo.Session, gd *discordgo.GuildDelete) {
	g := gd.Guild

	for i, wbg := range wb.guilds {
		if g.ID == wbg.ID {
			if i == 0 {
				wb.guilds = wb.guilds[1:]

				return
			}

			if i+1 == len(wb.guilds) {
				wb.guilds = wb.guilds[:len(wb.guilds)-1]

				return
			}

			lastIndex := len(wb.guilds) - 1

			wb.guilds[i] = wb.guilds[lastIndex]
			wb.guilds = wb.guilds[:lastIndex]

			return
		}
	}
}

// OnMessageCreate is the handler for the MESSAGE_CREATE event.
func (wb *Wirebot) OnMessageCreate(ds *discordgo.Session, m *discordgo.MessageCreate) {
	var (
		err error

		log = wb.logger()
	)

	if wb.Prefix == "" {
		return
	}
	if !strings.HasPrefix(m.Content, wb.Prefix) {
		return
	}

	var (
		head string
		tail string
	)

	parts := strings.SplitN(m.Content, " ", 2)
	if len(parts) == 0 {
		return
	}

	head = strings.TrimPrefix(parts[0], wb.Prefix)
	if len(parts) > 1 {
		tail = parts[1]
	}

	if tail == "" && strings.HasPrefix(head, wb.Prefix) {
		tail = strings.TrimPrefix(head, wb.Prefix)
		head = "emoji"

		if tail == "" {
			head = ""
		}
	}

	switch head {
	case "m":
		fallthrough
	case "mute":
		err = wb.cmdMute(ds, m, tail)

	case "r":
		fallthrough
	case "rule":
		err = wb.cmdRule(ds, m, tail)
	}

	if err != nil {
		log.Print(err)

		_, sendErr := ds.ChannelMessageSend(m.ChannelID, err.Error())
		if sendErr != nil {
			log.Print(sendErr)
		}
	}
}
