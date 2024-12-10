package jexample

import (
	"jdb"
	"jdebug"
	"jglobal"
	"jkcp"
	"jlog"
	"jmongo"
	"jnet"
	"jtcp"
	"jweb"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type DDD struct {
	Uin  uint32
	Name string
}

func Run() {
	// network()
	// mongo()
	// redis()
	event()
	schedule()
}

func network() {
	jnet.Tcp.Register(1, func(id uint64, pack *jtcp.Pack) {
		jlog.Debug(jdebug.StructToString(pack))
		jnet.Tcp.Send(id, 1, []byte("ok!"))
	})
	jnet.Kcp.Register(2, func(id uint64, pack *jkcp.Pack) {
		jlog.Debug(jdebug.StructToString(pack))
		jnet.Kcp.Send(id, 1, []byte("ok!"))
	})
	jnet.Web.Register(1, func(id uint64, pack *jweb.Pack) {
		jlog.Debug(jdebug.StructToString(pack))
		jnet.Kcp.Send(id, 1, []byte("ok!"))
	})
}

func mongo() {
	in := &jmongo.Input{
		Col:     "test",
		Filter:  bson.M{"uin": 1},
		Project: bson.M{"name": 1},
	}
	var ou any
	in = &jmongo.Input{
		Col: "test",
	}
	count, _ := jdb.Mongo.EstimatedDocumentCount(in)
	jlog.Debug(count)
	count, _ = jdb.Mongo.CountDocuments(in)
	jlog.Debug(count)
	ou = &DDD{}
	jdb.Mongo.FindOne(in, ou)
	jlog.Debug(ou)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 0},
		Sort:   bson.M{"uin": 1},
		Limit:  1,
	}
	ou = &[]*DDD{}
	jdb.Mongo.FindMany(in, ou)
	jlog.Debug(ou)
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
	res, _ := jdb.Redis.Do("SET", "jamger", "123", "EX", 3)
	jlog.Debug(res)
	res, _ = jdb.Redis.Do("GET", "jamger")
	jlog.Debug(reflect.TypeOf(res).Kind())
	scr := `
		local value = redis.call('GET', KEYS[1])
		return value
	`
	res, _ = jdb.Redis.DoScript(scr, []string{"jamger"})
	jlog.Debug(res)
}

func schedule() {
	id := jglobal.Schedule.DoEvery("* * * * * *", func() {
		jlog.Debug("doevery")
		jglobal.Event.Emit(jglobal.EVENT_TEST, nil)
	})

	jglobal.Schedule.DoAt(20*time.Second, func() {
		jlog.Debug("doat")
		jglobal.Schedule.Stop(id)
	})
}

func event() {
	jglobal.Event.Register(jglobal.EVENT_TEST, func(context any) {
		jlog.Debug("recv event test1")
	})
	jglobal.Event.Register(jglobal.EVENT_TEST, func(context any) {
		jlog.Debug("recv event test2")
	})
}
