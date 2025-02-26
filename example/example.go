package jexample

import (
	"errors"
	"jdb"
	"jevent"
	"jglobal"
	"jlog"
	"jmongo"
	"reflect"
	"time"

	"jschedule"

	"github.com/nsqio/go-nsq"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

type DDD struct {
	Uin  uint32
	Name string
}

type User struct {
	Id   int `gorm:"primaryKey"`
	Name string
	gorm.Model
}

// ------------------------- outside -------------------------

func Init() {
	// network()
	// mysql()
	// mongo()
	// redis()
	// schedule()
	// event()
	// rpc()
}

// ------------------------- inside -------------------------

func network() {
	// jnet.Tcp.Register(1, func(id uint64, pack *jtcp.Pack) {
	// 	jlog.Debug(jdebug.StructToString(pack))
	// 	jnet.Tcp.Send(id, 1, []byte("ok!"))
	// })
	// jnet.Kcp.Register(2, func(id uint64, pack *jkcp.Pack) {
	// 	jlog.Debug(jdebug.StructToString(pack))
	// 	jnet.Kcp.Send(id, 1, []byte("ok!"))
	// })
	// jnet.Web.Register(1, func(id uint64, pack *jweb.Pack) {
	// 	jlog.Debug(jdebug.StructToString(pack))
	// 	jnet.Kcp.Send(id, 1, []byte("ok!"))
	// })
}

func mysql() {
	var res *gorm.DB
	var n int64
	user := &User{
		Id:   8,
		Name: "ddd",
	}
	users := &[]User{
		{Id: 0, Name: "abc"},
		{Id: 2, Name: "kkk"},
	}
	users2 := &map[string]any{}
	var name string
	names := &[]string{}

	jdb.Mysql.Create(user)
	jdb.Mysql.Select("name").Create(user)
	jdb.Mysql.Omit("id").Create(user)
	jdb.Mysql.CreateInBatches(users, 2)

	jdb.Mysql.Table("users").Count(&n)

	res = jdb.Mysql.First(user)
	res = jdb.Mysql.Take(user)
	res = jdb.Mysql.Last(user)
	res = jdb.Mysql.Select("id", "name").Find(users)
	res = jdb.Mysql.Table("users").Take(user)
	res = jdb.Mysql.Model(user).Take(users2)
	res = jdb.Mysql.Where("name = ?", "kkk").First(user)
	res = jdb.Mysql.Not("name = ?", "abc").First(user)
	res = jdb.Mysql.Order("name desc").Find(users)
	res = jdb.Mysql.Limit(1).Offset(1).Find(users)
	res = jdb.Mysql.Group("name").Having("id = 1").Find(users)
	res = jdb.Mysql.Distinct("name").Find(users)
	res = jdb.Mysql.Table("users").Pluck("name", names)
	jdb.Mysql.OriginSql().Raw("select * from users").Scan(users)
	rows, _ := jdb.Mysql.OriginSql().Raw("select name from users").Rows()
	for rows.Next() {
		rows.Scan(&name)
		jdb.Mysql.ScanRows(rows, names)
	}

	res = jdb.Mysql.Table("users").Where("name = ?", "kkk").Update("name", "k")
	res = jdb.Mysql.Table("users").Where("name = ?", "k").Updates(map[string]any{"name": "kkk"})
	res = jdb.Mysql.Save(user)

	res = jdb.Mysql.Delete(user)
	if res != nil {
		jlog.Debug(res.RowsAffected)
		jlog.Debug(errors.Is(res.Error, gorm.ErrRecordNotFound))
	}
	jlog.Debug(n)
	jlog.Debug(name)
	jlog.Debug(names)
	jlog.Debug(user)
	jlog.Debug(users)
	jlog.Debug(users2)
}

func mongo() {
	in := &jmongo.Input{
		Col:     "test",
		Filter:  bson.M{"uin": 1},
		Project: bson.M{"name": 1},
	}
	var out any
	in = &jmongo.Input{
		Col: "test",
	}
	count, _ := jdb.Mongo.EstimatedDocumentCount(in)
	jlog.Debug(count)
	count, _ = jdb.Mongo.CountDocuments(in)
	jlog.Debug(count)
	out = &DDD{}
	jdb.Mongo.FindOne(in, out)
	jlog.Debug(out)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 0},
		Sort:   bson.M{"uin": 1},
		Limit:  1,
	}
	out = &[]*DDD{}
	jdb.Mongo.FindMany(in, out)
	jlog.Debug(out)
	in = &jmongo.Input{
		Col:    "test",
		Insert: &DDD{2, "2"},
	}
	jdb.Mongo.InsertOne(in)
	in = &jmongo.Input{
		Col: "test",
		InsertMany: []any{
			&DDD{Uin: 3, Name: "3"},
			&DDD{Uin: 4, Name: "4"},
		},
	}
	jdb.Mongo.InsertMany(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 2},
		Update: bson.M{"$set": bson.M{"name": "2"}},
	}
	jdb.Mongo.UpdateOne(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 2},
		Update: bson.M{"$set": bson.M{"name": "2"}},
	}
	jdb.Mongo.UpdateMany(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 4},
	}
	jdb.Mongo.DeleteOne(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 3},
	}
	jdb.Mongo.DeleteMany(in)
}

func redis() {
	rsp, _ := jdb.Redis.Do("SET", "jamger", "123", "EX", 3)
	jlog.Debug(rsp)
	rsp, _ = jdb.Redis.Do("GET", "jamger")
	jlog.Debug(reflect.TypeOf(rsp).Kind())
	scr := `
		local value = redis.call('GET', KEYS[1])
		return value
	`
	rsp, _ = jdb.Redis.DoScript(scr, []string{"jamger"})
	jlog.Debug(rsp)
}

func schedule() {
	id := jschedule.DoEvery(1*time.Second, func() {
		jlog.Debug("doevery")
	})
	jschedule.Stop(id)
	id = jschedule.DoCron("* * * * * *", func() {
		jlog.Debug("docron")
		jevent.Event.LocalPublish(jevent.EVENT_TEST_1, nil)
		if jglobal.SERVER == "jamger1" {
			jevent.Event.RemotePublish(jevent.EVENT_TEST_2, []byte("recv remote event"))
		}
	})
	jschedule.Stop(id)
	id = jschedule.DoAt(20*time.Second, func() {
		jlog.Debug("doat")
	})
	jschedule.Stop(id)
}

func event() {
	jevent.Event.LocalRegister(jevent.EVENT_TEST_1, func(context any) {
		jlog.Debug("recv local event")
	})
	jevent.Event.LocalRegister(jevent.EVENT_TEST_1, func(context any) {
		jlog.Debug("recv local event")
	})
	jevent.Event.RemoteRegister(jevent.EVENT_TEST_2, func(msg *nsq.Message) error {
		jlog.Debug(jglobal.SERVER, string(msg.Body))
		return nil
	})
}

func rpc() {
}
