package tcp

import (
	"encoding/binary"
	"io"
	"net"
)

// tcp stream structure:
// |   head   |   body   |   head   |   body   |
// | ----------------------------------------- |
// |    2     |   ...    |    2     |   ...    |

// pack structure:
// |         pack        |         pack        |
// | ----------------------------------------- |
// |   cmd    |   data   |   cmd    |   data   |
// | ----------------------------------------- |
// |    2     |   ...    |    2     |   ...    |

const (
	HEAD_SIZE = 2
	CMD_SIZE  = 2
)

type Pack struct {
	Cmd  uint16
	Data []byte
}

// ------------------------- outside -------------------------

func MakePack() Pack {
	return Pack{}
}

// ------------------------- package -------------------------

func recvPack(con net.Conn) (pack Pack, err error) {
	buffer := make([]byte, HEAD_SIZE)
	_, err = io.ReadFull(con, buffer)
	if err != nil {
		return
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	_, err = io.ReadFull(con, buffer)
	if err != nil {
		return
	}
	pack.Cmd = binary.LittleEndian.Uint16(buffer)
	pack.Data = buffer[CMD_SIZE:]
	return
}

func sendPack(con net.Conn, pack Pack) error {
	bodySize := CMD_SIZE + len(pack.Data)
	size := HEAD_SIZE + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[HEAD_SIZE:], pack.Cmd)
	copy(buffer[HEAD_SIZE+CMD_SIZE:], pack.Data)
	for pos := 0; pos < size; {
		n, err := con.Write(buffer)
		if err != nil {
			break
		}
		pos += n
	}
	return nil
}
