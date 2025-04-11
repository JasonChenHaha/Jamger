package jmongo

import (
	"context"
	"fmt"
	"jconfig"
	"jlog"
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
	Upsert     bool
	RetDoc     options.ReturnDocument
}

func (in *Input) String() string {
	return fmt.Sprintf("col:%s, filter:%s, insert:%s, insertmany:%s, update:%s", in.Col, in.Filter, in.Insert, in.InsertMany, in.Update)
}

type Mongo struct {
	*mongo.Client
	base string
	coll sync.Map
}

// ------------------------- outside -------------------------

func NewMongo(base string) *Mongo {
	mog := &Mongo{base: base}
	opts := options.Client().
		ApplyURI(jconfig.GetString("mongo.uri")).
		SetSocketTimeout(time.Duration(jconfig.GetInt("mongo.socketTimeout")) * time.Millisecond)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("connect to mongo")
	mog.Client = client
	return mog
}

func (mog *Mongo) EstimatedDocumentCount(in *Input) (int64, error) {
	co := mog.getCollection(in.Col)
	rsp, err := co.EstimatedDocumentCount(context.Background())
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return rsp, err
}

func (mog *Mongo) CountDocuments(in *Input) (int64, error) {
	co := mog.getCollection(in.Col)
	opts := options.Count()
	if in.Filter == nil {
		in.Filter = bson.M{}
		opts.SetHint("_id_")
	}
	rsp, err := co.CountDocuments(context.Background(), in.Filter, opts)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return rsp, err
}

// out需要为结构体指针 *(bson.M or struct)
func (mog *Mongo) FindOne(in *Input, out any) error {
	co := mog.getCollection(in.Col)
	opts := options.FindOne()
	opts.SetProjection(in.Project)
	err := co.FindOne(context.Background(), in.Filter, opts).Decode(out)
	if err != nil && err != mongo.ErrNoDocuments {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

// out需要为结构体指针 *(bson.M or struct)
func (mog *Mongo) FindOneAndUpdate(in *Input, out any) error {
	co := mog.getCollection(in.Col)
	opts := options.FindOneAndUpdate()
	opts.SetProjection(in.Project)
	opts.SetUpsert(in.Upsert)
	opts.SetReturnDocument(in.RetDoc)
	err := co.FindOneAndUpdate(context.Background(), in.Filter, in.Update, opts).Decode(out)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

// out需要为结构体指针切片的指针 *[]*(bson.M or struct)
func (mog *Mongo) FindMany(in *Input, out any) error {
	co := mog.getCollection(in.Col)
	if in.Filter == nil {
		in.Filter = bson.M{}
	}
	opts := options.Find()
	opts.SetProjection(in.Project)
	opts.SetSort(in.Sort)
	opts.SetLimit(in.Limit)
	cursor, err := co.Find(context.Background(), in.Filter, opts)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
		return err
	}
	err = cursor.All(context.Background(), out)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

func (mog *Mongo) InsertOne(in *Input) error {
	co := mog.getCollection(in.Col)
	_, err := co.InsertOne(context.Background(), in.Insert)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

func (mog *Mongo) InsertMany(in *Input) error {
	co := mog.getCollection(in.Col)
	_, err := co.InsertMany(context.Background(), in.InsertMany)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

func (mog *Mongo) UpdateOne(in *Input) error {
	co := mog.getCollection(in.Col)
	opts := options.Update()
	opts.SetUpsert(in.Upsert)
	_, err := co.UpdateOne(context.Background(), in.Filter, in.Update, opts)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

func (mog *Mongo) UpdateMany(in *Input) error {
	co := mog.getCollection(in.Col)
	opts := options.Update()
	opts.SetUpsert(in.Upsert)
	_, err := co.UpdateMany(context.Background(), in.Filter, in.Update, opts)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

func (mog *Mongo) DeleteOne(in *Input) error {
	co := mog.getCollection(in.Col)
	_, err := co.DeleteOne(context.Background(), in.Filter)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

func (mog *Mongo) DeleteMany(in *Input) error {
	co := mog.getCollection(in.Col)
	_, err := co.DeleteMany(context.Background(), in.Filter)
	if err != nil {
		jlog.Errorf("%s, %s", err, in)
	}
	return err
}

// ------------------------- inside -------------------------

func (mog *Mongo) getCollection(col string) *mongo.Collection {
	co, ok := mog.coll.Load(col)
	if !ok {
		co = mog.Database(mog.base).Collection(col)
		mog.coll.Store(col, co)
	}
	return co.(*mongo.Collection)
}
