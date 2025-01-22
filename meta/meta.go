package jmeta

import "jdb"

// ------------------------- outside -------------------------

func Init() {
	tmp := []interface{}{}
	for _, v := range tmp {
		jdb.Mysql.AutoMigrate(v)
	}
}
