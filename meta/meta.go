package jmeta

import "jdb"

// ------------------------- inside -------------------------

func Init() {
	tmp := []interface{}{}
	for _, v := range tmp {
		jdb.Mysql.AutoMigrate(v)
	}
}
