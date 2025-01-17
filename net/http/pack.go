package jhttp

import (
	"encoding/binary"
	"jpb"
)

// pack structure:
// |         pack        |         pack        |
// +----------+----------+----------+----------+
// |   cmd    |   data   |   cmd    |   data   |
// +----------+----------+----------+----------+
// |    2     |   ...    |    2     |   ...    |

const (
	CmdSize = 2
)

type Pack struct {
	Cmd  jpb.CMD
	Data []byte
}

// ------------------------- package -------------------------

func SerializePack(pack *Pack) []byte {
	msg := make([]byte, CmdSize+len(pack.Data))
	binary.LittleEndian.PutUint16(msg, uint16(pack.Cmd))
	copy(msg[CmdSize:], pack.Data)
	return msg
}

func UnserializeToPack(msg []byte) *Pack {
	return &Pack{
		Cmd:  jpb.CMD(binary.LittleEndian.Uint16(msg)),
		Data: msg[CmdSize:],
	}
}
