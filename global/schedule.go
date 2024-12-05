package jglobal

import (
	jlog "jamger/log"
	"time"

	"github.com/robfig/cron/v3"
)

var Schedule *schedule

type schedule struct {
	cron *cron.Cron
}

// ------------------------- inside -------------------------

func init() {
	Schedule = &schedule{cron: cron.New(cron.WithSeconds())}
	Schedule.cron.Start()
}

// ------------------------- outside -------------------------

// 定时t时间后触发cmd
func (sch *schedule) DoAt(t time.Duration, cmd func()) any {
	timer := time.NewTimer(t)
	go func() {
		<-timer.C
		cmd()
	}()
	return timer
}

// 定时固定间隔触发cmd
func (sch *schedule) DoEvery(format string, cmd func()) any {
	id, err := sch.cron.AddFunc(format, cmd)
	if err != nil {
		jlog.Panic(err)
	}
	return id
}

func (sch *schedule) Stop(id any) {
	switch v := id.(type) {
	case cron.EntryID:
		sch.cron.Remove(v)
	case *time.Timer:
		v.Stop()
	}
}
