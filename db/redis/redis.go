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
	return re
}

func (re *Redis) Exist(key string) (bool, error) {
	rsp, err := re.client.Exists(context.Background(), key).Result()
	if err != nil {
		jlog.Error(err)
		return false, err
	}
	return rsp > 0, err
}

func (re *Redis) Set(key string, value any, expire int64) (string, error) {
	rsp, err := re.client.Set(context.Background(), key, value, time.Duration(expire)*time.Second).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) Get(key string) (string, error) {
	rsp, err := re.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		err = nil
	} else if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) HSet(key string, values ...any) (int64, error) {
	rsp, err := re.client.HSet(context.Background(), key, values...).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) HDel(key string, fields ...string) (int64, error) {
	rsp, err := re.client.HDel(context.Background(), key, fields...).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) HGet(key string, field string) (string, error) {
	rsp, err := re.client.HGet(context.Background(), key, field).Result()
	if err == redis.Nil {
		err = nil
	} else if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) HGetAll(key string) (map[string]string, error) {
	rsp, err := re.client.HGetAll(context.Background(), key).Result()
	if err == redis.Nil {
		err = nil
	} else if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) Del(key string) (int64, error) {
	rsp, err := re.client.Del(context.Background(), key).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) Do(args ...any) (any, error) {
	rsp, err := re.client.Do(context.Background(), args...).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) DoScript(script string, keys []string, args ...any) (any, error) {
	rsp, err := re.client.Eval(context.Background(), script, keys, args...).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}

func (re *Redis) Flush() (any, error) {
	rsp, err := re.client.FlushAll(context.Background()).Result()
	if err != nil {
		jlog.Error(err)
	}
	return rsp, err
}
