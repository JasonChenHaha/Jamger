package jconfig

import (
	jlog "jamger/log"
	"os"

	"github.com/spf13/viper"
)

var g_cfg = viper.New()

func init() {
	path, err := os.Getwd()
	if err != nil {
		jlog.Panic(err)
	}
	g_cfg.AddConfigPath(path)
	g_cfg.SetConfigName("config")
	g_cfg.SetConfigType("yml")
	if err = g_cfg.ReadInConfig(); err != nil {
		jlog.Panic(err)
	}
}

func Get(key string) any {
	return g_cfg.Get(key)
}
