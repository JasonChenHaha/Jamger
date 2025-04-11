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
	Data []*jpb.Address
}

var Addr *Address

// ------------------------- outside -------------------------

func Init() {
	Addr = &Address{Data: []*jpb.Address{}}
	in := &jmongo.Input{
		Col: jglobal.MONGO_ADDRESS,
	}
	out := bson.A{}
	if err := jdb.Mongo.FindMany(in, &out); err != nil {
		return
	}
	for _, v := range out {
		ma, err := bson.Marshal(v.(primitive.D))
		if err != nil {
			jlog.Error(err)
			return
		}
		addr := bson.M{}
		err = bson.Unmarshal(ma, &addr)
		if err != nil {
			jlog.Error(err)
			return
		}
		Addr.Data = append(Addr.Data, &jpb.Address{
			Name:      addr["name"].(string),
			Addr:      addr["address"].(string),
			Longitude: addr["longitude"].(float64),
			Latitude:  addr["latitude"].(float64),
		})
	}
}
