package jredis

import (
	"context"
	"jconfig"
	"jlog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

// ------------------------- outside -------------------------

func NewRedis() *Redis {
	re := &Redis{}
	client := redis.NewClient(&redis.Options{
		Addr:         jconfig.GetString("redis.addr"),
		Password:     jconfig.GetString("redis.password"),
		ReadTimeout:  time.Duration(jconfig.GetInt("redis.rTimeout")) * time.Millisecond,
		WriteTimeout: time.Duration(jconfig.GetInt("redis.wTimeout")) * time.Millisecond,
	})
	jlog.Info("connect to redis")
	re.client = client
	re.Flush()
	return re
}

func (re *Redis) Exist(key string) (bool, error) {
	rsp, err := re.client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return rsp > 0, err
}

func (re *Redis) HSet(key string, values ...any) (int64, error) {
	return re.client.HSet(context.Background(), key, values...).Result()
}

func (re *Redis) HGet(key string, field string) (string, error) {
	rsp, err := re.client.HGet(context.Background(), key, field).Result()
	if err == redis.Nil {
		err = nil
	}
	return rsp, err
}

func (re *Redis) HGetAll(key string) (map[string]string, error) {
	rsp, err := re.client.HGetAll(context.Background(), key).Result()
	if err == redis.Nil {
		err = nil
	}
	return rsp, err
}

func (re *Redis) Del(key string) (int64, error) {
	return re.client.Del(context.Background(), key).Result()
}

func (re *Redis) Do(args ...any) (any, error) {
	return re.client.Do(context.Background(), args...).Result()
}

func (re *Redis) DoScript(script string, keys []string, args ...any) (any, error) {
	return re.client.Eval(context.Background(), script, keys, args...).Result()
}

func (re *Redis) Flush() (any, error) {
	return re.client.FlushAll(context.Background()).Result()
}
