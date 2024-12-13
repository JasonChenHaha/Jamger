package jdb

import (
	"jconfig"
	"jmongo"
	"jmysql"
	"jredis"
)

// ------------------------- outside -------------------------

var Mysql *jmysql.Jmysql
var Mongo *jmongo.Mongo
var Redis *jredis.Redis

func Init() {
	if jconfig.Get("mysql") != nil {
		Mysql = jmysql.NewMysql()
	}
	if jconfig.Get("mongo") != nil {
		Mongo = jmongo.NewMongo("boomboat")
	}
	if jconfig.Get("redis") != nil {
		Redis = jredis.NewRedis()
	}
}
