package jdb

import (
	"jconfig"
	"jmongo"
	"jmysql"
	"jredis"
)

var Mysql *jmysql.Jmysql
var Mongo *jmongo.Mongo
var Redis *jredis.Redis

// ------------------------- outside -------------------------

func Init() {
	if jconfig.Get("mysql") != nil {
		Mysql = jmysql.NewMysql()
	}
	if jconfig.Get("mongo") != nil {
		Mongo = jmongo.NewMongo(jconfig.GetString("mongo.db"))
	}
	if jconfig.Get("redis") != nil {
		Redis = jredis.NewRedis()
	}
}
