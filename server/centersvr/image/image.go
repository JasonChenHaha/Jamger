package jimage

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

const (
	EXPIRE = 60
)

type Img struct {
	cache *jglobal.TimeCache[uint32, []byte]
}

var Image *Img

// ------------------------- outside -------------------------

func Init() {
	Image = &Img{cache: jglobal.NewTimeCache[uint32, []byte](EXPIRE)}
}

// 图片压缩
func (img *Img) Compress(source []byte) ([]byte, error) {
	// 解析图片
	r := bytes.NewReader(source)
	img2, format, err := image.Decode(r)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	// 调整尺寸
	img3 := imaging.Resize(img2, 250, 0, imaging.Lanczos)
	buffer := &bytes.Buffer{}
	switch format {
	case "jpeg":
		if err = jpeg.Encode(buffer, img3, &jpeg.Options{Quality: 50}); err != nil {
			jlog.Error(err)
			return nil, err
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err = encoder.Encode(buffer, img3); err != nil {
			jlog.Error(err)
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

// 生成图片id
// func (img *Img) GenUid() (int64, error) {
// 	in := &jmongo.Input{
// 		Col:     jglobal.MONGO_IMAGE,
// 		Filter:  bson.M{"_id": int64(0)},
// 		Update:  bson.M{"$inc": bson.M{"idc": int64(1)}},
// 		Upsert:  true,
// 		RetDoc:  options.After,
// 		Project: bson.M{"idc": 1},
// 	}
// 	out := bson.M{}
// 	if err := jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
// 		return 0, err
// 	}
// 	return out["idc"].(int64), nil
// }

// 添加图片
func (img *Img) Add(uid uint32, data []byte) error {
	img.cache.Set(uid, data)
	in := &jmongo.Input{
		Col:    jglobal.MONGO_IMAGE,
		Filter: bson.M{"_id": uint64(uid)},
		Update: bson.M{"$set": bson.M{"image": data}},
		Upsert: true,
	}
	return jdb.Mongo.UpdateOne(in)
}

// 获得图片(带缓存)
func (img *Img) Get(uid uint32) ([]byte, error) {
	if data := img.cache.Get(uid); data != nil {
		return data, nil
	} else {
		in := &jmongo.Input{
			Col:     jglobal.MONGO_IMAGE,
			Filter:  bson.M{"_id": uint64(uid)},
			Project: bson.M{"image": 1},
		}
		out := bson.M{}
		if err := jdb.Mongo.FindOne(in, &out); err != nil {
			return nil, err
		}
		img.cache.Set(uid, data)
		return out["image"].(primitive.Binary).Data, nil
	}
}
