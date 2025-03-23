package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// ------------------------- outside -------------------------

func Init() {}

// 判断账号是否存在
func IsUserExist(id string) (bson.M, error) {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"basic.id": id},
		Project: bson.M{"_id": 1, "basic.admin": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOne(in, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// 密码加密
func EncryptPwd(pwd string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
}

// 生成用户uid
func GenUserUid() (uint32, error) {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": int64(0)},
		Update:  bson.M{"$inc": bson.M{"uidc": int64(1)}},
		Upsert:  true,
		RetDoc:  options.After,
		Project: bson.M{"uidc": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
		return 0, err
	}
	return uint32(out["uidc"].(int64)), nil
}

// 创建用户
func CreateUser(uid uint32, id string, secret []byte) error {
	in := &jmongo.Input{
		Col:    jglobal.MONGO_USER,
		Insert: bson.M{"_id": uid, "basic": bson.M{"id": id, "pwd": secret}},
	}
	return jdb.Mongo.InsertOne(in)
}

func WxCreateUser(uid uint32, id string) error {
	in := &jmongo.Input{
		Col:    jglobal.MONGO_USER,
		Insert: bson.M{"_id": uid, "basic": bson.M{"id": id}},
	}
	return jdb.Mongo.InsertOne(in)
}

// 账号校验
func AccountCheck(id string, pwd string) (uint32, error) {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"basic.id": id},
		Project: bson.M{"basic.pwd": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOne(in, &out); err != nil {
		return 0, err
	}
	secret := out["basic"].(bson.M)["pwd"].(primitive.Binary)
	if err := bcrypt.CompareHashAndPassword(secret.Data, []byte(pwd)); err != nil {
		// 密码错误
		jlog.Error(err)
		return 0, nil
	}
	return uint32(out["_id"].(int64)), nil
}
