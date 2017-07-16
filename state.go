package wirebot

import (
	"errors"
	"os"
)

type State struct {
	File  string `json:"-"`
	Token string `json:"token"`

	BannedWords     map[GuildID][]string                       `json:"banned_words"`
	GuildRules      map[GuildID][]string                       `json:"guild_rules"`
	GuildTags       map[GuildID]map[string]string              `json:"guild_tags"`
	MutedUsers      map[GuildID][]*UserMute                    `json:"muted_users"`
	UserKudos       map[GuildID][]*UserKudos                   `json:"user_kudos"`
	UserPermissions map[GuildID]map[UserID]map[Permission]bool `json:"user_permissions"`
	UserWarnings    map[GuildID][]*UserWarning                 `json:"user_warnings"`
}

// Sync persists the state to the filesystem.
func (s *State) Sync() error {
	if s.File == "" {
		return errors.New("missing path to file")
	}

	f, err := os.OpenFile(s.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	err = f.Sync()

	f.Close()

	return err
}
