package jaddress

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	data *jglobal.TimeCache[uint32, []*jpb.Address]
}

var Addr *Address

// ------------------------- outside -------------------------

func Init() {
	Addr = &Address{data: jglobal.NewTimeCache[uint32, []*jpb.Address](86400)}
}

func (addr *Address) Get() []*jpb.Address {
	if addrs := addr.data.Get(0); addrs != nil {
		return addrs
	} else {
		in := &jmongo.Input{
			Col: jglobal.MONGO_ADDRESS,
		}
		out := bson.A{}
		if err := jdb.Mongo.FindMany(in, &out); err != nil {
			return nil
		}
		tmp := []*jpb.Address{}
		for _, v := range out {
			ma, err := bson.Marshal(v.(primitive.D))
			if err != nil {
				jlog.Error(err)
				return nil
			}
			out := bson.M{}
			err = bson.Unmarshal(ma, &out)
			if err != nil {
				jlog.Error(err)
				return nil
			}
			tmp = append(tmp, &jpb.Address{
				Name:      out["name"].(string),
				Addr:      out["address"].(string),
				Longitude: out["longitude"].(float64),
				Latitude:  out["latitude"].(float64),
			})
			addr.data.Set(0, tmp)
		}
		return tmp
	}
}
