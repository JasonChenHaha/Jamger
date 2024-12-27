package jmysql

// GORM document
// https://learnku.com/docs/gorm/v2

import (
	"context"
	"database/sql"
	"jconfig"
	"jlog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Jmysql struct {
	db       *gorm.DB
	base     bool
	rTimeout time.Duration
	sTimeout time.Duration
}

// ------------------------- inside -------------------------

func (ms *Jmysql) clone() *Jmysql {
	return &Jmysql{
		rTimeout: ms.rTimeout,
		sTimeout: ms.sTimeout,
	}
}

// ------------------------- outside -------------------------

func NewMysql() *Jmysql {
	ms := &Jmysql{
		base:     true,
		rTimeout: time.Duration(jconfig.GetInt("mysql.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("mysql.sTimeout")) * time.Millisecond,
	}
	st := &schema.NamingStrategy{
		SingularTable: false, // 开启表名复数形式
		NoLowerCase:   false, // 开启自动转小写
	}
	lo := logger.New(jlog.Logger(), logger.Config{
		SlowThreshold: time.Duration(jconfig.GetInt("mysql.slowThreshold")) * time.Millisecond,
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
	ms.db = db
	return ms
}

func (ms *Jmysql) OriginSql() *gorm.DB {
	return ms.db
}

func (ms *Jmysql) Table(name string, args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Table(name, args...)
	return o
}

func (ms *Jmysql) Select(query any, args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Select(query, args...)
	return o
}

func (ms *Jmysql) Model(value any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Model(value)
	return o
}

func (ms *Jmysql) Omit(columns ...string) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Omit(columns...)
	return o
}

func (ms *Jmysql) Where(query any, args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Where(query, args...)
	return o
}

func (ms *Jmysql) Not(query any, args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Not(query, args...)
	return o
}

func (ms *Jmysql) Order(value any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Order(value)
	return o
}

func (ms *Jmysql) Limit(limit int) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Limit(limit)
	return o
}

func (ms *Jmysql) Offset(offset int) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Offset(offset)
	return o
}

func (ms *Jmysql) Group(name string) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Group(name)
	return o
}

func (ms *Jmysql) Having(query any, args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Having(query, args...)
	return o
}

func (ms *Jmysql) Distinct(args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Distinct(args...)
	return o
}

func (ms *Jmysql) Joins(query string, args ...any) *Jmysql {
	o := ms
	if ms.base {
		o = ms.clone()
	}
	o.db = ms.db.Joins(query, args...)
	return o
}

// 自动建表
func (ms *Jmysql) AutoMigrate(dst ...any) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	ms.db.WithContext(ctx).AutoMigrate(dst...)
}

// 计数
func (ms *Jmysql) Count(count *int64) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Count(count)
}

// 写
func (ms *Jmysql) Create(value any) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	ms.db.WithContext(ctx).Create(value)
}

// 批量写
func (ms *Jmysql) CreateInBatches(value any, batchSize int) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	ms.db.WithContext(ctx).CreateInBatches(value, batchSize)
}

// 查首条(主键升序)
func (ms *Jmysql) First(dest any, conds ...any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).First(dest, conds...)
}

// 查一条
func (ms *Jmysql) Take(dest any, conds ...any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Take(dest)
}

// 查尾条(逐渐降序)
func (ms *Jmysql) Last(dest any, conds ...any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Last(dest, conds...)
}

// 查找
func (ms *Jmysql) Find(dest any, conds ...any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Find(dest, conds...)
}

// 查找
func (ms *Jmysql) Scan(dest any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Scan(dest)
}

// 批次查找
func (ms *Jmysql) FindInBatches(dest any, batchSize int, f func(*gorm.DB, int) error) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).FindInBatches(dest, batchSize, f)
}

// 查询列
func (ms *Jmysql) Pluck(column string, dest any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.rTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Pluck(column, dest)
}

// 原生查找
func (ms *Jmysql) ScanRows(rows *sql.Rows, sql any) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).ScanRows(rows, sql)
}

// 更新
func (ms *Jmysql) Update(column string, value any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Update(column, value)
}

// 多个更新
func (ms *Jmysql) Updates(values any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Updates(values)
}

// 全量更新
func (ms *Jmysql) Save(value any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Save(value)
}

// 删除
func (ms *Jmysql) Delete(value any, conds ...any) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), ms.sTimeout)
	defer cancel()
	return ms.db.WithContext(ctx).Delete(value, conds...)
}
