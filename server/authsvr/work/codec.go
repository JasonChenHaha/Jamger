package jwork

import (
	"encoding/binary"
	"jglobal"
	"jpb"
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
	binary.LittleEndian.PutUint16(raw[uidSize+gateSize:], uint16(pack.Cmd))
	copy(raw[uidSize+gateSize+cmdSize:], data)
	pack.Data = raw
	return nil
}

func rpcDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize+gateSize:]))
	pack.Data = raw[uidSize+gateSize+cmdSize:]
	return nil
}
