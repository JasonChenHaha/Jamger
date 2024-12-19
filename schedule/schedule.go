package jschedule

// schedule内定时器在协程内触发，所有cmd需要注意并发安全

import (
	"jlog"
	"time"

	"github.com/robfig/cron/v3"
)

var sch *schedule

type schedule struct {
	cron *cron.Cron
}

// ------------------------- inside -------------------------

func init() {
	sch = &schedule{cron: cron.New(cron.WithSeconds())}
	sch.cron.Start()
}

// ------------------------- outside -------------------------

// 定时t时间后触发cmd
func DoAt(t time.Duration, cmd func()) any {
	timer := time.NewTimer(t)
	go func() {
		<-timer.C
		cmd()
	}()
	return timer
}

// 定时固定间隔触发cmd
func DoEvery(format string, cmd func()) any {
	id, err := sch.cron.AddFunc(format, cmd)
	if err != nil {
		jlog.Panic(err)
	}
	return id
}

func Stop(id any) {
	switch v := id.(type) {
	case cron.EntryID:
		sch.cron.Remove(v)
	case *time.Timer:
		v.Stop()
	}
}
