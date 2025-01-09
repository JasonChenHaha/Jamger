package jweb

import (
	"encoding/binary"
	pb "jpb"
)

// websocket pack structure:
// |         pack        |         pack        |
// +----------+----------+----------+----------+
// |   cmd    |   data   |   cmd    |   data   |
// +----------+----------+----------+----------+
// |    2     |   ...    |    2     |   ...    |

const (
	CmdSize = 2
)

type Pack struct {
	Cmd  pb.CMD
	Data []byte
}

// ------------------------- package -------------------------

func makePack(cmd pb.CMD, data []byte) *Pack {
	return &Pack{
		Cmd:  cmd,
		Data: data,
	}
}

func unserializeData(data []byte) *Pack {
	return &Pack{
		Cmd:  pb.CMD(binary.LittleEndian.Uint16(data)),
		Data: data[CmdSize:],
	}
}

func serializePack(pack *Pack) []byte {
	size := CmdSize + len(pack.Data)
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(pack.Cmd))
	copy(buffer[CmdSize:], pack.Data)
	return buffer
}
