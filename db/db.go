package jdb

import (
	jconfig "jamger/config"
	jmysql "jamger/db/mysql"
)

// ------------------------- outside -------------------------

var Mysql *jmysql.Mysql

func Run() {
	if cfg := jconfig.Get("mysql"); cfg != nil {
		Mysql = jmysql.NewMysql()
		Mysql.Run()
	}
}
