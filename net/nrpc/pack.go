package jnrpc

import (
	"encoding/binary"
	"jglobal"
	"jpb"
)

// pack structure:
// |         pack        |
// +----------+----------+
// |   cmd    |   data   |
// +----------+----------+
// |    2     |   ...    |

const (
	CmdSize = 2
)

// ------------------------- package -------------------------

func encodePack(pack *jglobal.Pack) {
	data := pack.Data.([]byte)
	raw := make([]byte, CmdSize+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[CmdSize:], data)
	pack.Data = raw
}

func decodeToPack(pack *jglobal.Pack, raw []byte) {
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
	pack.Data = raw[CmdSize:]
}
