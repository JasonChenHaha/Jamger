package jmysql

// GORM document
// https://learnku.com/docs/gorm/v2

import (
	jconfig "jamger/config"
	jlog "jamger/log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	*gorm.DB
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
	sqlDb, err := db.DB()
	if err != nil {
		jlog.Fatal(err)
	}
	sqlDb.SetMaxOpenConns(jconfig.GetInt("mysql.maxOpenCon"))
	sqlDb.SetMaxIdleConns(jconfig.GetInt("mysql.maxIdleCon"))
	sqlDb.SetConnMaxLifetime(time.Duration(jconfig.GetInt("mysql.maxLifeTime")) * time.Millisecond)
	sqlDb.SetConnMaxIdleTime(time.Duration(jconfig.GetInt("mysql.maxIdleTime")) * time.Millisecond)
	ms.DB = db
}
