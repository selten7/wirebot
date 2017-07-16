package wirebot

import (
	"time"
)

func (wb *Wirebot) jobUnmute() error {
	if wb.State == nil {
		return nil
	}

	var (
		log = wb.logger()
		now = time.Now()
	)

	for guildID, mutedUsers := range wb.State.MutedUsers {
		for _, mutedUser := range mutedUsers {
			if now.After(mutedUser.MutedUntil) {
				log.Printf("unmuting %s on guild %s", mutedUser.UserID, guildID)

				err := wb.unmuteUser(string(guildID), mutedUser.UserID)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}

	return nil
}
