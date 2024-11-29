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

type Input struct {
	Col        string
	Filter     any
	Insert     any
	InsertMany []any
	Update     any
	Sort       any
	Limit      int64
	Project    any
}

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

func (mog *Mongo) EstimatedDocumentCount(in *Input) (int64, error) {
	co := mog.GetCollection(in.Col)
	return co.EstimatedDocumentCount(context.Background())
}

func (mog *Mongo) CountDocuments(in *Input) (int64, error) {
	co := mog.GetCollection(in.Col)
	opts := options.Count()
	if in.Filter == nil {
		in.Filter = bson.M{}
		opts.SetHint("_id_")
	}
	return co.CountDocuments(context.Background(), in.Filter, opts)
}

// out需要为结构体指针 *struct
func (mog *Mongo) FindOne(in *Input, out any) error {
	co := mog.GetCollection(in.Col)
	opts := options.FindOne()
	opts.SetProjection(in.Project)
	return co.FindOne(context.Background(), in.Filter, opts).Decode(out)
}

// out需要为结构体指针切片的指针 *[]*struct
func (mog *Mongo) FindMany(in *Input, out any) error {
	co := mog.GetCollection(in.Col)
	opts := options.Find()
	opts.SetProjection(in.Project)
	opts.SetSort(in.Sort)
	opts.SetLimit(in.Limit)
	cursor, err := co.Find(context.Background(), in.Filter, opts)
	if err != nil {
		return err
	}
	return cursor.All(context.Background(), out)
}

func (mog *Mongo) InsertOne(in *Input) error {
	co := mog.GetCollection(in.Col)
	_, err := co.InsertOne(context.Background(), in.Insert)
	return err
}

func (mog *Mongo) InsertMany(in *Input) error {
	co := mog.GetCollection(in.Col)
	_, err := co.InsertMany(context.Background(), in.InsertMany)
	return err
}

func (mog *Mongo) UpdateOne(in *Input) error {
	co := mog.GetCollection(in.Col)
	_, err := co.UpdateOne(context.Background(), in.Filter, in.Update)
	return err
}

func (mog *Mongo) UpdateMany(in *Input) error {
	co := mog.GetCollection(in.Col)
	_, err := co.UpdateMany(context.Background(), in.Filter, in.Update)
	return err
}

func (mog *Mongo) DeleteOne(in *Input) error {
	co := mog.GetCollection(in.Col)
	_, err := co.DeleteOne(context.Background(), in.Filter)
	return err
}

func (mog *Mongo) DeleteMany(in *Input) error {
	co := mog.GetCollection(in.Col)
	_, err := co.DeleteMany(context.Background(), in.Filter)
	return err
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
