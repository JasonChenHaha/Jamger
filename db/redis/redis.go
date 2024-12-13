package jredis

import (
	"context"
	"jconfig"
	"jlog"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

// ------------------------- outside -------------------------

func NewRedis() *Redis {
	re := &Redis{}
	client := redis.NewClient(&redis.Options{
		Addr:     jconfig.GetString("redis.addr"),
		Password: jconfig.GetString("redis.password"),
	})
	jlog.Info("connect to redis")
	re.client = client
	return re
}

func (re *Redis) Do(args ...any) (any, error) {
	return re.client.Do(context.Background(), args...).Result()
}

func (re *Redis) DoScript(script string, keys []string, args ...any) (any, error) {
	return re.client.Eval(context.Background(), script, keys, args...).Result()
}
