package jconfig

import (
	"fmt"
	"jlog"
	"jtrash"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var config = viper.New()

// ------------------------- inside -------------------------

func init() {
	index := strings.LastIndex(os.Args[1], "/")
	config.AddConfigPath(os.Args[1][:index])
	config.SetConfigName(os.Args[1][index+1:])
	config.SetConfigType("yml")
	if err := config.ReadInConfig(); err != nil {
		jlog.Panic(err)
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
				if num, ok := jtrash.TransTimeStrToUint64(o); ok {
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
