package jtcp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"jglobal"
	"jtrash"
	"net"
)

// tcp stream structure:
// |   head   |   pack   |   head   |   pack   |
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

func recvPack(con net.Conn) (pack *Pack, err error) {
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
	err = decryptPack(pack)
	return pack, err
}

func decryptPack(pack *Pack) error {
	var data []byte
	var err error
	if pack.Cmd == jglobal.CMD_SIGN_UP_REQ {
		data, err = jtrash.RSADecrypt(jglobal.RSA_PRIVATE_KEY, pack.Data)
		if err != nil {
			return err
		}
		size := len(data) - 4
		if binary.LittleEndian.Uint32(data[size:]) != crc32.ChecksumIEEE(data[:size]) {
			return fmt.Errorf("checksum failed. cmd: %d", pack.Cmd)
		}
		data = data[:size]
	} else {
		data = pack.Data
	}
	pack.Data = data
	return nil
}

func sendPack(con net.Conn, pack *Pack) error {
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
