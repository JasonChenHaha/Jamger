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

// ------------------------- outside -------------------------

func Init() {
	sch = &schedule{cron: cron.New(cron.WithSeconds())}
	sch.cron.Start()
}

// 定时t时间后触发cmd
func DoAt(t time.Duration, cmd func()) any {
	timer := time.NewTimer(t)
	go func() {
		<-timer.C
		cmd()
	}()
	return timer
}

func DoEvery(t time.Duration, cmd func()) any {
	ticker := time.NewTicker(t)
	go func() {
		for range ticker.C {
			cmd()
		}
	}()
	return ticker
}

// 定时固定间隔触发cmd
func DoCron(format string, cmd func()) any {
	id, err := sch.cron.AddFunc(format, cmd)
	if err != nil {
		jlog.Panic(err)
	}
	return id
}

func Stop(id any) {
	switch v := id.(type) {
	case *time.Timer:
		v.Stop()
	case *time.Ticker:
		v.Stop()
	case cron.EntryID:
		sch.cron.Remove(v)
	}
}
