package jconfig

import (
	jlog "jamger/log"
	"os"

	"github.com/spf13/viper"
)

var g_cfg = viper.Viper{}

func init() {
	path, err := os.Getwd()
	if err != nil {
		jlog.Panic(err)
	}
	g_cfg.AddConfigPath(path)
	g_cfg.SetConfigName("config")
	g_cfg.SetConfigType("yml")
	err = g_cfg.ReadInConfig()
	if err != nil {
		jlog.Panic(err)
	}
}

func Get(key string) any {
	return g_cfg.Get(key)
}
