package jglobal

import (
	"jlog"
	"regexp"
	"strconv"
	"time"
)

// 将带时间后缀的值，全部转成以毫秒单位的数值
func TransTimeStrToUint64(str string) (uint64, bool) {
	var scale uint64
	re := regexp.MustCompile(`[a-zA-Z]+$`)
	u := re.FindString(str)
	switch u {
	case "h":
		scale = 3600000
	case "m":
		scale = 60000
	case "s":
		scale = 1000
	case "ms":
		scale = 1
	default:
		return 0, false
	}
	num, err := strconv.ParseUint(str[:len(str)-len(u)], 10, 64)
	if err != nil {
		jlog.Fatal(err)
	}
	return num * scale, true
}

// 今天的0点时间
func GetTodayZeroTime() *time.Time {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return &zero
}

// 明天的0点时间
func GetTomorrowZeroTime() *time.Time {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return &zero
}

// 到下秒的时间
func TimeToSecond() time.Duration {
	now := time.Now()
	return time.Second - time.Duration(now.Nanosecond())
}

// 到下分钟的时间
func TimeToMinute() time.Duration {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	return zero.Sub(now)
}

// 到明天的时间
func TimeToTomorrow() time.Duration {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return zero.Sub(now)
}
