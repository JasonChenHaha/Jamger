package jwork

import (
	"encoding/binary"
	"fmt"
	"jglobal"
	"jpb"
	"juser"
)

const (
	uidSize      = 4
	cmdSize      = 2
	aesKeySize   = 16
	checksumSize = 4
)

// ------------------------- package -------------------------

// server http pack structure:
//       +--------------------+
//       |        pack        |
//       +---------+----------+
// aes ( |   cmd   |   data   | )
//       +---------+----------+
//       |    2    |   ...    |
//       +---------+----------+

func rpcEncode(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, uidSize+cmdSize+len(data))
	if pack.User != nil {
		user := pack.User.(*juser.User)
		user.Unlock()
		binary.LittleEndian.PutUint32(raw, uint32(user.Uid))
	}
	binary.LittleEndian.PutUint16(raw[uidSize:], uint16(pack.Cmd))
	copy(raw[uidSize+cmdSize:], data)
	pack.Data = raw
	return nil
}

func rpcDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	if pack.User == nil && uid != 0 {
		user := juser.GetUser(uid)
		if user == nil {
			return fmt.Errorf("no such user, uid = %d", uid)
		}
		user.Lock()
		pack.User = user
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize:]))
	pack.Data = raw[uidSize+cmdSize:]
	return nil
}
