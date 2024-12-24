package jtrash

import (
	"jlog"
	"os"
	"os/signal"
	"time"
)

// ------------------------- outside -------------------------

func Rcover() {
	if r := recover(); r != nil {
		jlog.Panic(r)
	}
}

// 阻塞进程
func Keep() {
	mainC := make(chan os.Signal, 1)
	signal.Notify(mainC, os.Interrupt)
	<-mainC
}

// 今天0点的时刻
func GetTodayZeroTime() *time.Time {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return &zero
}

// 明天0点的时间
func GetTomorrowZeroTime() *time.Time {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return &zero
}

// 到下秒的时间差
func TimeToSecond() time.Duration {
	now := time.Now()
	return time.Second - time.Duration(now.Nanosecond())
}

// 到下分钟的时间差
func TimeToMinute() time.Duration {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	return zero.Sub(now)
}

// 到明天0点的时间差
func TimeToTomorrow() time.Duration {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return zero.Sub(now)
}

// 到指定时刻的时间差
func TimeToTime(hour int) time.Duration {
	now := time.Now()
	time5 := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if now.After(time5) {
		time5 = time5.Add(24 * time.Hour)
	}
	return time5.Sub(now)
}
