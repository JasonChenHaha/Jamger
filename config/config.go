package jconfig

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var config = viper.New()

// ------------------------- inside -------------------------

func Init() {
	index := strings.LastIndex(os.Args[1], "/")
	config.AddConfigPath(os.Args[1][:index])
	config.SetConfigName(os.Args[1][index+1:])
	config.SetConfigType("yml")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	formatCfg()
}

// 格式化配置中的数值
func formatCfg() {
	var fu func(path string, cfg map[string]any)
	fu = func(path string, cfg map[string]any) {
		for k, v := range cfg {
			switch o := v.(type) {
			case string:
				if num, err := transTimeStrToUint64(o); err == nil {
					if len(path) == 0 {
						config.Set(k, num)
					} else {
						config.Set(fmt.Sprintf("%s.%s", path, k), num)
					}
				}
			case map[string]any:
				if len(path) == 0 {
					fu(k, o)
				} else {
					fu(fmt.Sprintf("%s.%s", path, k), o)
				}
			}
		}
	}
	fu("", config.AllSettings())
}

// 将带时间后缀的值，全部转成以毫秒单位的数值
func transTimeStrToUint64(str string) (uint64, error) {
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
		return 0, fmt.Errorf("suffix is invalid")
	}
	num, err := strconv.ParseUint(str[:len(str)-len(u)], 10, 64)
	return num * scale, err
}

// ------------------------- outside -------------------------

func Get(key string) any {
	return config.Get(key)
}

func GetInt(key string) int {
	return config.GetInt(key)
}

func GetString(key string) string {
	return config.GetString(key)
}

func GetBool(key string) bool {
	return config.GetBool(key)
}
