package jkcp

import (
	"encoding/binary"
	"io"
	"jpb"

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

type Pack struct {
	Cmd  jpb.CMD
	Data []byte
}

const (
	HeadSize = 2
	CmdSize  = 2
)

// ------------------------- package -------------------------

func recvPack(con *kcp.UDPSession) (pack *Pack, err error) {
	buffer := make([]byte, HeadSize)
	if _, err = io.ReadFull(con, buffer); err != nil {
		return
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	if _, err = io.ReadFull(con, buffer); err != nil {
		return
	}
	pack = &Pack{
		Cmd:  jpb.CMD(binary.LittleEndian.Uint16(buffer)),
		Data: buffer[CmdSize:],
	}
	return
}

func sendPack(con *kcp.UDPSession, pack *Pack) error {
	bodySize := CmdSize + len(pack.Data)
	size := HeadSize + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[HeadSize:], uint16(pack.Cmd))
	copy(buffer[HeadSize+CmdSize:], pack.Data)
	for pos := 0; pos < size; {
		n, err := con.Write(buffer)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
