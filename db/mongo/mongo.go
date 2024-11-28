package jmongo

import (
	"context"
	jconfig "jamger/config"
	jlog "jamger/log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	*mongo.Client
	base string
	coll sync.Map
}

// ------------------------- outside -------------------------

func NewMongo(base string) *Mongo {
	return &Mongo{base: base}
}

func (mog *Mongo) Run() {
	opts := options.Client().
		ApplyURI(jconfig.GetString("mongo.uri")).
		SetTimeout(time.Duration(jconfig.GetInt("mongo.timeout")) * time.Millisecond).
		SetMaxPoolSize(uint64(jconfig.GetInt("mongo.maxPoolSize"))).
		SetMaxConnIdleTime(time.Duration(jconfig.GetInt("mong.maxIdleTime")) * time.Millisecond)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("connect to mongo")
	mog.Client = client
}

// out需要为结构体指针 *struct
func (mog *Mongo) FindOne(col string, filter bson.D, out any) error {
	co := mog.GetCollection(col)
	return co.FindOne(context.Background(), filter).Decode(out)
}

// out需要为结构体指针切片的指针 *[]*struct
func (mog *Mongo) FindMany(col string, filter bson.D, out any) error {
	co := mog.GetCollection(col)
	cursor, err := co.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	return cursor.All(context.Background(), out)
}

// in需要为结构体指针 *struct
func (mog *Mongo) InsertOne(col string, in any) error {
	co := mog.GetCollection(col)
	_, err := co.InsertOne(context.Background(), in)
	return err
}

func (mog *Mongo) InsertMany(col string, in any) {
	co := mog.GetCollection(col)
}

// ------------------------- inside -------------------------

func (mog *Mongo) GetCollection(col string) *mongo.Collection {
	co, ok := mog.coll.Load(col)
	if !ok {
		co = mog.Database(mog.base).Collection(col)
		mog.coll.Store(col, co)
	}
	return co.(*mongo.Collection)
}
