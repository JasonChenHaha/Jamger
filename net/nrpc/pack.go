package jnrpc

import (
	"encoding/binary"
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

type Pack struct {
	cmd  jpb.CMD
	data []byte
}

// ------------------------- package -------------------------

func encodeFromPack(pack *Pack) []byte {
	msg := make([]byte, CmdSize+len(pack.data))
	binary.LittleEndian.PutUint16(msg, uint16(pack.cmd))
	copy(msg[CmdSize:], pack.data)
	return msg
}

func decodeToPack(msg []byte) *Pack {
	return &Pack{
		cmd:  jpb.CMD(binary.LittleEndian.Uint16(msg)),
		data: msg[CmdSize:],
	}
}
