package jmysql

// GORM document
// https://learnku.com/docs/gorm/v2

import (
	jconfig "jamger/config"
	jlog "jamger/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	db *gorm.DB
}

func NewMysql() *Mysql {
	return &Mysql{}
}

func (ms *Mysql) Run() {
	db, err := gorm.Open(mysql.Open(jconfig.GetString("mysql.dsn")), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("connect to mysql")
	ms.db = db
}

func (ms *Mysql) AutoMigrate(tab interface{}) {
	ms.db.AutoMigrate(tab)
}
