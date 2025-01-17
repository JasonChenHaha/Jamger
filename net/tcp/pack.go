package jtcp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"jglobal"
	"jpb"
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

// sign up or sign in:
// client: cmd + rsa(payload + aeskey + checksum) -> server
// server: aes(payload + checksum) -> client

// other:
// client: aes(payload + checksum) -> server
// server: aes(payload + checksum) -> client

const (
	HeadSize     = 2
	CmdSize      = 2
	ChecksumSize = 4
	AesKeySize   = 16
)

type Pack struct {
	Cmd  jpb.CMD
	Data []byte
}

// ------------------------- package -------------------------

func recvPack(ses *Ses) (pack *Pack, err error) {
	// 读数据
	buffer := make([]byte, HeadSize)
	if _, err = io.ReadFull(ses.con, buffer); err != nil {
		return
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	if _, err = io.ReadFull(ses.con, buffer); err != nil {
		return
	}
	pack = &Pack{
		Cmd:  jpb.CMD(binary.LittleEndian.Uint16(buffer)),
		Data: buffer[CmdSize:],
	}
	return
}

func sendPack(ses *Ses, pack *Pack) error {
	bodySize := CmdSize + len(pack.Data)
	size := HeadSize + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[HeadSize:], uint16(pack.Cmd))
	copy(buffer[HeadSize+CmdSize:], pack.Data)
	// 写数据
	for pos := 0; pos < size; {
		n, err := ses.con.Write(buffer)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}

func parseRSAPack(pack *Pack) ([]byte, error) {
	if err := jglobal.RSADecrypt(jglobal.RSA_PRIVATE_KEY, &pack.Data); err != nil {
		return nil, err
	}
	size := len(pack.Data) - ChecksumSize
	code := pack.Data[size:]
	pack.Data = pack.Data[:size]
	if binary.LittleEndian.Uint32(code) != crc32.ChecksumIEEE(pack.Data) {
		return nil, fmt.Errorf("checksum failed, %d", pack.Cmd)
	}
	size -= AesKeySize
	aesKey := pack.Data[size:]
	pack.Data = pack.Data[:size]
	return aesKey, nil
}

func makeAESPack(ses *Ses, pack *Pack) error {
	size := len(pack.Data)
	pack.Data = append(pack.Data, make([]byte, 4)...)
	binary.LittleEndian.PutUint32(pack.Data[size:], crc32.ChecksumIEEE(pack.Data[:size]))
	return jglobal.AESEncrypt(ses.aesKey, &pack.Data)
}

func parseAESPack(aesKey []byte, pack *Pack) error {
	if err := jglobal.AESDecrypt(aesKey, &pack.Data); err != nil {
		return err
	}
	size := len(pack.Data) - ChecksumSize
	code := pack.Data[size:]
	pack.Data = pack.Data[:size]
	if binary.LittleEndian.Uint32(code) != crc32.ChecksumIEEE(pack.Data) {
		return fmt.Errorf("checksum failed, %d", pack.Cmd)
	}
	return nil
}
