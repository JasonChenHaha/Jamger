package juser2

import (
	"fmt"
	"jglobal"
	"jlog"
	"jpb"

	"go.mongodb.org/mongo-driver/bson"
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
		tmp := make([]*jpb.Record, len(v.(bson.M)))
		for k2, v2 := range v.(bson.M) {
			v3 := v2.(bson.M)
			tmp[jglobal.Atoi[int](k2)] = &jpb.Record{
				Score:     uint32(v3["score"].(int64)),
				Timestamp: uint64(v3["timestamp"].(int64)),
			}
		}
		re.Data = tmp
	}
}

// ------------------------- outside -------------------------

// 新增记录
func (re *Record) AddRecord(record *jpb.Record) {
	re.Data = append(re.Data, record)
	re.user.Lock()
	re.user.DirtyMongo[fmt.Sprintf("record.%d", len(re.Data)-1)] = record
	re.user.UnLock()
}

// 修改记录
func (re *Record) ModifyRecord(index uint32, record *jpb.Record) bool {
	if index >= uint32(len(re.Data)) {
		jlog.Errorf("index(%d) out of bounds(%d)", index, len(re.Data))
		return false
	}
	record.Timestamp = re.Data[index].Timestamp
	re.Data[index] = record
	re.user.Lock()
	re.user.DirtyMongo[fmt.Sprintf("record.%d", index)] = re.Data[index]
	re.user.UnLock()
	return true
}

// 删除记录
func (re *Record) DeleteRecord(index uint32) bool {
	if index >= uint32(len(re.Data)) {
		jlog.Errorf("index(%d) out of bounds(%d)", index, len(re.Data))
		return false
	}
	jglobal.SliceDeletePos(&re.Data, index)
	return true
}
