package jdb

import (
	jconfig "jamger/config"
	jmongo "jamger/db/mongo"
	jmysql "jamger/db/mysql"
)

// ------------------------- outside -------------------------

var Mysql *jmysql.Mysql
var Mongo *jmongo.Mongo

func Run() {
	if jconfig.Get("mysql") != nil {
		Mysql = jmysql.NewMysql()
		Mysql.Run()
	}
	if jconfig.Get("mongo") != nil {
		Mongo = jmongo.NewMongo("boomboat")
		Mongo.Run()
	}
}
