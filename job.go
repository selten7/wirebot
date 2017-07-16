package wirebot

import (
	"time"
)

type job struct {
	LastRun  time.Time
	RunEvery time.Duration
	Func     func() error
}
