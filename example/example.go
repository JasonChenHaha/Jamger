package jexample

import (
	jdb "jamger/db"
	jmongo "jamger/db/mongo"
	jlog "jamger/log"

	"go.mongodb.org/mongo-driver/bson"
)

type DDD struct {
	Uin  uint32
	Name string
}

func Run() {
	mongo()
}

func mongo() {
	in := &jmongo.Input{
		Col:     "test",
		Filter:  bson.M{"uin": 1},
		Project: bson.M{"name": 1},
	}
	var ou any
	// in = &jmongo.Input{
	// 	Col: "test",
	// }
	// count, err := jdb.Mongo.EstimatedDocumentCount(in)
	// if err != nil {
	// 	jlog.Error(err)
	// }
	// jlog.Debug(count)
	// count, err = jdb.Mongo.CountDocuments(in)
	// if err != nil {
	// 	jlog.Error(err)
	// }
	// jlog.Debug(count)
	// ou = &DDD{}
	// if err := jdb.Mongo.FindOne(in, ou); err != nil {
	// 	jlog.Error(err)
	// }
	// jlog.Debug(ou)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 0},
		Sort:   bson.M{"uin": 1},
		Limit:  1,
	}
	ou = &[]*DDD{}
	if err := jdb.Mongo.FindMany(in, ou); err != nil {
		jlog.Error(err)
	}
	jlog.Debug(ou)
	// in = &jmongo.Input{
	// 	Col:    "test",
	// 	Insert: &DDD{2, "2"},
	// }
	// if err := jdb.Mongo.InsertOne(in); err != nil {
	// 	jlog.Error(err)
	// }
	// in = &jmongo.Input{
	// 	Col: "test",
	// 	InsertMany: []any{
	// 		&DDD{Uin: 3, Name: "3"},
	// 		&DDD{Uin: 4, Name: "4"},
	// 	},
	// }
	// if err := jdb.Mongo.InsertMany(in); err != nil {
	// 	jlog.Error(err)
	// }
	// in = &jmongo.Input{
	// 	Col:    "test",
	// 	Filter: bson.M{"uin": 2},
	// 	Update: bson.M{"$set": bson.M{"name": "2"}},
	// }
	// if err := jdb.Mongo.UpdateOne(in); err != nil {
	// 	jlog.Error(err)
	// }
	// in = &jmongo.Input{
	// 	Col:    "test",
	// 	Filter: bson.M{"uin": 2},
	// 	Update: bson.M{"$set": bson.M{"name": "2"}},
	// }
	// if err := jdb.Mongo.UpdateMany(in); err != nil {
	// 	jlog.Error(err)
	// }
	// in = &jmongo.Input{
	// 	Col:    "test",
	// 	Filter: bson.M{"uin": 4},
	// }
	// if err := jdb.Mongo.DeleteOne(in); err != nil {
	// 	jlog.Error(err)
	// }
	// in = &jmongo.Input{
	// 	Col:    "test",
	// 	Filter: bson.M{"uin": 3},
	// }
	// if err := jdb.Mongo.DeleteMany(in); err != nil {
	// 	jlog.Error(err)
	// }
}
