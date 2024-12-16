package jconfig

import (
	"fmt"
	"jlog"
	"jtrash"
	"os"

	"github.com/spf13/viper"
)

var g_cfg = viper.New()

// ------------------------- inside -------------------------

func init() {
	path, err := os.Getwd()
	if err != nil {
		jlog.Panic(err)
	}
	g_cfg.AddConfigPath(path)
	g_cfg.SetConfigName(os.Args[1])
	g_cfg.SetConfigType("yml")
	if err = g_cfg.ReadInConfig(); err != nil {
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
						g_cfg.Set(k, num)
					} else {
						g_cfg.Set(fmt.Sprintf("%s.%s", path, k), num)
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
	fu("", g_cfg.AllSettings())
}

// ------------------------- outside -------------------------

func Get(key string) any {
	return g_cfg.Get(key)
}

func GetInt(key string) int {
	return g_cfg.GetInt(key)
}

func GetString(key string) string {
	return g_cfg.GetString(key)
}

func GetBool(key string) bool {
	return g_cfg.GetBool(key)
}
