package jkcp

import (
	"encoding/binary"
	"io"

	"github.com/xtaci/kcp-go"
)

// kcp stream structure:
// |   head   |   body   |   head   |   body   |
// +----------+----------+----------+----------+
// |    2     |   ...    |    2     |   ...    |

// pack structure:
// |         pack        |         pack        |
// +----------+----------+----------+----------+
// |   cmd    |   data   |   cmd    |   data   |
// +----------+----------+----------+----------+
// |    2     |   ...    |    2     |   ...    |

const (
	gHeadSize = 2
	gCmdSize  = 2
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

func recvPack(con *kcp.UDPSession) (pack *Pack, err error) {
	buffer := make([]byte, gHeadSize)
	if _, err = io.ReadFull(con, buffer); err != nil {
		return
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	if _, err = io.ReadFull(con, buffer); err != nil {
		return
	}
	pack = &Pack{
		Cmd:  binary.LittleEndian.Uint16(buffer),
		Data: buffer[gCmdSize:],
	}
	return
}

func sendPack(con *kcp.UDPSession, pack *Pack) error {
	bodySize := gCmdSize + len(pack.Data)
	size := gHeadSize + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[gHeadSize:], pack.Cmd)
	copy(buffer[gHeadSize+gCmdSize:], pack.Data)
	for pos := 0; pos < size; {
		n, err := con.Write(buffer)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
