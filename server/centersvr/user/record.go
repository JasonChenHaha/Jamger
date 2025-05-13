package juser2

import (
	"fmt"
	"jdb"
	"jglobal"
	"jlog"
	"jmedia"
	"jmongo"
	"jpb"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Record struct {
	user *User
	Data []*jpb.Record
}

// ------------------------- package -------------------------

func newRecord(user *User) *Record {
	return &Record{user: user}
}

func (re *Record) load(data bson.M) {
	if v, ok := data["record"]; ok {
		tmp := []*jpb.Record{}
		for _, v2 := range v.(bson.M) {
			if v3, ok := v2.(bson.M); ok {
				tmp = append(tmp, &jpb.Record{
					Uid:       uint32(v3["uid"].(int64)),
					Score:     uint32(v3["score"].(int64)),
					Muid:      uint32(v3["muid"].(int64)),
					Timestamp: uint64(v3["timestamp"].(int64)),
				})
			}
		}
		sort.Slice(tmp, func(i, j int) bool { return tmp[i].Timestamp < tmp[j].Timestamp })
		re.Data = tmp
	}
}

// ------------------------- outside -------------------------

// 新增记录
func (re *Record) AddRecord(record *jpb.Record) error {
	if record.Media != nil {
		uids, err := jmedia.Media.Add([]*jpb.Media{record.Media}, true)
		if err != nil {
			return err
		}
		record.Media = nil
		for k := range uids {
			record.Muid = k
		}
	}
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": int64(re.user.Uid)},
		Update:  bson.M{"$inc": bson.M{"record.ruidc": int64(1)}},
		Upsert:  true,
		RetDoc:  options.After,
		Project: bson.M{"record.ruidc": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
		return err
	}
	record.Uid = uint32(out["record"].(bson.M)["ruidc"].(int64))
	re.Data = append(re.Data, record)
	re.user.Lock()
	re.user.DirtyMongo[fmt.Sprintf("record.%d", record.Uid)] = record
	re.user.UnLock()
	return nil
}

// 修改记录
func (re *Record) ModifyRecord(index uint32, record *jpb.Record) error {
	if index >= uint32(len(re.Data)) {
		jlog.Errorf("index(%d) out of bounds(%d)", index, len(re.Data))
		return fmt.Errorf("index(%d) out of bounds(%d)", index, len(re.Data))
	}
	re.Data[index].Score = record.Score
	re.user.Lock()
	re.user.DirtyMongo[fmt.Sprintf("record.%d", re.Data[index].Uid)] = re.Data[index]
	re.user.UnLock()
	return nil
}

// 删除记录
func (re *Record) DeleteRecord(index uint32) error {
	if index >= uint32(len(re.Data)) {
		jlog.Errorf("index(%d) out of bounds(%d)", index, len(re.Data))
		return fmt.Errorf("index(%d) out of bounds(%d)", index, len(re.Data))
	}
	if re.Data[index].Muid != 0 {
		if err := jmedia.Media.Delete([]uint32{re.Data[index].Muid}); err != nil {
			return err
		}
	}
	re.user.Lock()
	re.user.DirtyMongo[fmt.Sprintf("record.%d", re.Data[index].Uid)] = nil
	re.user.UnLock()
	jglobal.SliceDeletePos(&re.Data, index)
	return nil
}
