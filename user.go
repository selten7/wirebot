package wirebot

import (
	"time"
)

type UserID string

type UserMute struct {
	UserID     string    `json:"user_id"`
	MutedUntil time.Time `json:"muted_until"`
}

type UserWarning struct {
	UserID    string `json:"user_id"`
	RuleIndex string `json:"rule_number"`
	Count     int    `json:"count"`
}

type UserKudos struct {
	UserID string `json:"user_id"`
	Kudos  int    `json:"kudos"`
}
