package jmeta

import jdb "jamger/db"

// ------------------------- outside -------------------------

func Init() {
	tmp := []interface{}{
		&Mail{},
		&User{},
	}
	for _, v := range tmp {
		jdb.Mysql.AutoMigrate(v)
	}
}
