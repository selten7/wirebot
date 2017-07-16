package wirebot

import (
	"errors"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Wirebot struct {
	Prefix string
	Logger Logger
	Token  string

	session *discordgo.Session
	guilds  []*discordgo.Guild

	StateMutex sync.Mutex
	State      *State

	muxJobs       sync.Mutex
	nextJobID     int
	jobs          map[int]*job
	jobsInterrupt chan struct{}
}

func (wb *Wirebot) logger() Logger {
	if wb.Logger == nil {
		return discardLogger
	}

	return wb.Logger
}

func (wb *Wirebot) addDefaultJobs() {
	wb.JobAdd(1*time.Minute, wb.jobUnmute)
}

func (wb *Wirebot) JobAdd(runEvery time.Duration, jobFunc func() error) int {
	wb.muxJobs.Lock()
	defer wb.muxJobs.Unlock()

	if wb.jobs == nil {
		wb.jobs = make(map[int]*job)
		wb.muxJobs.Unlock()

		wb.addDefaultJobs()

		wb.muxJobs.Lock()
	}

	id := wb.nextJobID
	wb.nextJobID++

	wb.jobs[id] = &job{
		RunEvery: runEvery,
		Func:     jobFunc,
	}

	wb.logger().Printf("registered job %d", id)

	return id
}

func (wb *Wirebot) JobDelete(id int) {
	wb.muxJobs.Lock()
	defer wb.muxJobs.Unlock()

	delete(wb.jobs, id)
}

func (wb *Wirebot) scheduleJobs() error {
	wb.muxJobs.Lock()
	defer wb.muxJobs.Unlock()

	if wb.jobs == nil {
		wb.jobs = make(map[int]*job)
		wb.muxJobs.Unlock()

		wb.addDefaultJobs()

		wb.muxJobs.Lock()
	}

	wb.logger().Print("scheduling jobs")

	if wb.jobsInterrupt != nil {
		return errors.New("jobs already running")
	}

	wb.jobsInterrupt = make(chan struct{})

	go func() {
		for {
			select {
			case <-wb.jobsInterrupt:
				return
			case <-time.After(1 * time.Minute):
				wb.runAllJobs()
			}
		}
	}()

	return nil
}

func (wb *Wirebot) runAllJobs() {
	wb.muxJobs.Lock()
	defer wb.muxJobs.Unlock()

	wb.StateMutex.Lock()
	defer wb.StateMutex.Unlock()

	var (
		err error
	)

	now := time.Now()
	for _, j := range wb.jobs {
		if now.After(j.LastRun.Add(j.RunEvery)) {
			err = j.Func()
			if err != nil {
				wb.logger().Print(err)
			}
		}
	}
}

// Open starts a connection with Discord.
func (wb *Wirebot) Open() error {
	if wb.session != nil {
		return errors.New("session already opened")
	}

	var (
		err error
		ds  *discordgo.Session

		log = wb.logger()
	)

	ds, err = discordgo.New("Bot " + wb.Token)
	if err != nil {
		return err
	}

	//ds.AddHandler(wb.OnGuildCreate)
	//ds.AddHandler(wb.OnGuildDelete)
	ds.AddHandler(wb.OnMessageCreate)
	ds.AddHandler(wb.OnReady)

	/*
	   wb.passiveActions = []PassiveAction{
	   passivePixiv,
	   wb.passiveWordBan,
	   }
	*/

	log.Print("opening session")

	err = ds.Open()
	if err != nil {
		return err
	}

	log.Print("session opened successfully")

	wb.session = ds

	wb.scheduleJobs()

	return nil
}

// Close terminates the connection with Discord.
func (wb *Wirebot) Close() error {
	if wb.session == nil {
		return errors.New("no session")
	}

	log := wb.logger()
	log.Print("closing session")

	err := wb.session.Close()

	return err
}

func (wb *Wirebot) channelGuild(channelID string) (guildID string, ok bool) {
	if len(wb.guilds) == 0 {
		return
	}

	for _, g := range wb.guilds {
		for _, c := range g.Channels {
			if c.ID == channelID {
				guildID = g.ID
				ok = true

				return
			}
		}
	}

	return
}
