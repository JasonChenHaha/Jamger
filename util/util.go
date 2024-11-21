package util

import (
	jlog "jamger/log"
	"regexp"
	"strconv"
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
