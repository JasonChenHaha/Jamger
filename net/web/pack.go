package jweb

import "encoding/binary"

// websocket pack structure:
// |         pack        |         pack        |
// +----------+----------+----------+----------+
// |   cmd    |   data   |   cmd    |   data   |
// +----------+----------+----------+----------+
// |    2     |   ...    |    2     |   ...    |

const (
	gCmdSize = 2
)

type Pack struct {
	Cmd  uint16
	Data []byte
}

// ------------------------- package -------------------------

func makePack(cmd uint16, data []byte) *Pack {
	return &Pack{
		Cmd:  cmd,
		Data: data,
	}
}

func unserializeData(data []byte) *Pack {
	return &Pack{
		Cmd:  binary.LittleEndian.Uint16(data),
		Data: data[gCmdSize:],
	}
}

func serializePack(pack *Pack) []byte {
	size := gCmdSize + len(pack.Data)
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, pack.Cmd)
	copy(buffer[gCmdSize:], pack.Data)
	return buffer
}
