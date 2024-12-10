package jmysql

// GORM document
// https://learnku.com/docs/gorm/v2

import (
	"jconfig"
	"jlog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Jmysql struct {
	*gorm.DB
}

// ------------------------- outside -------------------------

func NewMysql() *Jmysql {
	return &Jmysql{}
}

func (ms *Jmysql) Run() {
	st := &schema.NamingStrategy{
		SingularTable: false, // 开启表名复数形式
		NoLowerCase:   false, // 开启自动转小写
	}
	lo := logger.New(jlog.GetLog(), logger.Config{
		SlowThreshold: time.Duration(jconfig.GetInt("slowThreshold")) * time.Millisecond,
		LogLevel:      logger.Warn,
	})
	db, err := gorm.Open(mysql.Open(jconfig.GetString("mysql.dsn")), &gorm.Config{
		NamingStrategy:         st,
		Logger:                 lo,
		SkipDefaultTransaction: true, // 禁用事务
	})
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("connect to mysql")
	ms.DB = db
}
