package jconfig

import (
	"os"

	"github.com/spf13/viper"
)

var g_cfg *viper.Viper

func Load() error {
	// jlog.Println("config init...")

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	// jlog.Println(path)

	g_cfg = viper.New()
	g_cfg.AddConfigPath(path)
	g_cfg.SetConfigName("config")
	g_cfg.SetConfigType("conf")

	g_cfg.ReadInConfig()
	return nil
}
