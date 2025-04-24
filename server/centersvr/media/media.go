package jmedia

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"jconfig"
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jpb"

	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Medi struct {
	data *jglobal.TimeCache[uint32, *jpb.Media]
}

const (
	MEDIA_IMAGE = 1
	MEDIA_VIDEO = 2
)

var Media *Medi

// ------------------------- outside -------------------------

func Init() {
	Media = &Medi{data: jglobal.NewTimeCache[uint32, *jpb.Media](int64(jconfig.GetInt("media.expire")))}
}

// 添加媒体
func (me *Medi) Add(medias []*jpb.Media) (map[uint32]uint32, error) {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_MEDIA,
		Filter:  bson.M{"_id": int64(0)},
		Update:  bson.M{"$inc": bson.M{"muidc": int64(len(medias))}},
		Upsert:  true,
		RetDoc:  options.After,
		Project: bson.M{"muidc": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
		return nil, err
	}
	uid, uids := uint32(out["muidc"].(int64))-uint32(len(medias))+1, map[uint32]uint32{}
	many := []any{}
	for _, v := range medias {
		me.data.Set(uid, v)
		if v.Video == nil {
			// 添加图片
			uids[uid] = MEDIA_IMAGE
			many = append(many, bson.M{"_id": uid, "image": v.Image})
			jlog.Infof("upload image %d", len(v.Image))
		} else {
			// 添加视频(和预览图片)
			uids[uid] = MEDIA_VIDEO
			many = append(many, bson.M{"_id": uid, "image": v.Image, "video": v.Video})
			jlog.Infof("upload video %d", len(v.Video))
		}
		uid++
	}
	in = &jmongo.Input{
		Col:        jglobal.MONGO_MEDIA,
		InsertMany: many,
	}
	return uids, jdb.Mongo.InsertMany(in)
}

// 修改媒体
func (me *Medi) Modify(uid uint32, media *jpb.Media) error {
	in := &jmongo.Input{
		Col:    jglobal.MONGO_MEDIA,
		Filter: bson.M{"_id": uid},
	}
	if media.Video == nil {
		in.Update = bson.M{"$set": bson.M{"image": media.Image}, "$unset": bson.M{"video": ""}}
	} else {
		in.Update = bson.M{"$set": bson.M{"image": media.Image, "video": media.Video}}
	}
	return jdb.Mongo.UpdateOne(in)
}

// 获得图片(带缓存)
func (me *Medi) GetImage(uid uint32) ([]byte, error) {
	if media := me.data.Get(uid); media != nil {
		return media.Image, nil
	} else {
		in := &jmongo.Input{
			Col:     jglobal.MONGO_MEDIA,
			Filter:  bson.M{"_id": uint64(uid)},
			Project: bson.M{"image": 1, "video": 1},
		}
		out := bson.M{}
		if err := jdb.Mongo.FindOne(in, &out); err != nil {
			return nil, err
		}
		media = &jpb.Media{Image: out["image"].(primitive.Binary).Data}
		if video, ok := out["video"].(primitive.Binary); ok {
			media.Video = video.Data
		}
		me.data.Set(uid, media)
		return media.Image, nil
	}
}

// 获得视频(带缓存)
func (me *Medi) GetVideo(uid uint32) ([]byte, error) {
	if media := me.data.Get(uid); media != nil {
		return media.Video, nil
	} else {
		in := &jmongo.Input{
			Col:     jglobal.MONGO_MEDIA,
			Filter:  bson.M{"_id": uint64(uid)},
			Project: bson.M{"image": 1, "video": 1},
		}
		out := bson.M{}
		if err := jdb.Mongo.FindOne(in, &out); err != nil {
			return nil, err
		}
		media = &jpb.Media{}
		if v, ok := out["image"]; ok {
			media.Image = v.(primitive.Binary).Data
		}
		if v, ok := out["video"]; ok {
			media.Video = v.(primitive.Binary).Data
		}
		me.data.Set(uid, media)
		return media.Video, nil
	}
}

// 删除媒体
func (me *Medi) Delete(uids []uint32) error {
	for _, uid := range uids {
		me.data.Del(uid)
	}
	in := &jmongo.Input{
		Col:    jglobal.MONGO_MEDIA,
		Filter: bson.M{"_id": bson.M{"$in": uids}},
	}
	return jdb.Mongo.DeleteMany(in)
}

// ------------------------- inside -------------------------

// 图片压缩
func (me *Medi) compressImage(source []byte) ([]byte, error) {
	// 解析图片
	r := bytes.NewReader(source)
	img2, format, err := image.Decode(r)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	// 调整尺寸
	if resize := jconfig.GetInt("media.image.resize"); resize != 0 {
		img2 = imaging.Resize(img2, 250, 0, imaging.Lanczos)
	}
	buffer := &bytes.Buffer{}
	switch format {
	case "jpeg":
		if err = jpeg.Encode(buffer, img2, &jpeg.Options{Quality: jconfig.GetInt("media.image.quality")}); err != nil {
			jlog.Error(err)
			return nil, err
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err = encoder.Encode(buffer, img2); err != nil {
			jlog.Error(err)
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}
