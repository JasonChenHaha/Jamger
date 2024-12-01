package jdb

import (
	jconfig "jamger/config"
	jmongo "jamger/db/mongo"
	jmysql "jamger/db/mysql"
	jredis "jamger/db/redis"
)

// ------------------------- outside -------------------------

var Mysql *jmysql.Mysql
var Mongo *jmongo.Mongo
var Redis *jredis.Redis

func Run() {
	if jconfig.Get("mysql") != nil {
		Mysql = jmysql.NewMysql()
		Mysql.Run()
	}
	if jconfig.Get("mongo") != nil {
		Mongo = jmongo.NewMongo("boomboat")
		Mongo.Run()
	}
	if jconfig.Get("redis") != nil {
		Redis = jredis.NewRedis()
		Redis.Run()
	}
}
