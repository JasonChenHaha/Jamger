package jwork

import (
	"encoding/binary"
	"fmt"
	"jglobal"
	"jpb"
	"juser"
)

const (
	uidSize  = 4
	gateSize = 4
	cmdSize  = 2
)

// ------------------------- package -------------------------

// rpc pack head structure:
// +------------------------------+
// |             head             |
// +---------+----------+---------+
// |   uid   |   gate   |   cmd   |
// +---------+----------+---------+
// |    4    |    4     |    2    |
// +------------------------------+
// rpc pack structure:
// +---------------------+
// |        pack         |
// +----------+----------+
// |   head   |   data   |
// +----------+----------+
// |    ..    |    ..    |
// +----------+----------+

func rpcEncode(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, uidSize+gateSize+cmdSize+len(data))
	if pack.Ctx != nil {
		user := pack.Ctx.(*juser.User)
		binary.LittleEndian.PutUint32(raw, uint32(user.Uid))
		binary.LittleEndian.PutUint32(raw[uidSize:], uint32(user.Gate))
	}
	binary.LittleEndian.PutUint16(raw[uidSize+gateSize:], uint16(pack.Cmd))
	copy(raw[uidSize+gateSize+cmdSize:], data)
	pack.Data = raw
	return nil
}

func rpcDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	if pack.Ctx == nil && uid != 0 {
		user := juser.GetUser(uid)
		if user == nil {
			return fmt.Errorf("no such user, uid(%d)", uid)
		}
		pack.Ctx = user
	}
	if pack.Ctx != nil {
		pack.Ctx.(*juser.User).SetGate(int(binary.LittleEndian.Uint32(raw[uidSize:])))
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize+gateSize:]))
	pack.Data = raw[uidSize+gateSize+cmdSize:]
	return nil
}
