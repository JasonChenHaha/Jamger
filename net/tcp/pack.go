package jtcp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"jglobal"
	"jpb"
	"juser"
)

// client pack structure:
// +--------------------------------------------------------------------+
// |                               pack                                 |
// +----------+---------+-------+---------+----------+--------------+---+
// |   size   |   uid   |       |   cmd   |   data   |   checksum   |   |
// +----------+---------+ aes ( +---------+----------+--------------+ ) |
// |    2     |    4    |       |    2    |   ...    |      4       |   |
// +----------+---------+-------+---------+----------+--------------+---+

// server pack structure:
// +-------------------------------------------+
// |                   pack                    |
// +----------+-------+---------+----------+---+
// |   size   |       |   cmd   |   data   |   |
// +----------+ aes ( +---------+----------+ ) |
// |    2     |       |    2    |   size   |   |
// +----------+-------+---------+----------+---+

const (
	packSize     = 2
	uidSize      = 4
	cmdSize      = 2
	checksumSize = 4
)

// ------------------------- package -------------------------

// 接收、解密数据
func recvAndDecodeToPack(pack *jglobal.Pack, ses *Ses) error {
	raw := make([]byte, packSize)
	if _, err := io.ReadFull(ses.con, raw); err != nil {
		return err
	}
	size := binary.LittleEndian.Uint16(raw)
	raw = make([]byte, size)
	if _, err := io.ReadFull(ses.con, raw); err != nil {
		return err
	}
	if ses.aesKey == nil {
		uid := binary.LittleEndian.Uint32(raw)
		user := juser.GetUser(uid)
		ses.aesKey = user.AesKey
	}
	raw = raw[uidSize:]
	if err := jglobal.AesDecrypt(ses.aesKey, &raw); err != nil {
		return err
	}
	posChecksum := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[posChecksum:]) != crc32.ChecksumIEEE(raw[:posChecksum]) {
		return fmt.Errorf("checksum failed")
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[:cmdSize]))
	pack.Data = raw[cmdSize:]
	return nil
}

// 加密、发送数据
func encodeAndSendPack(pack *jglobal.Pack, ses *Ses) error {
	data := pack.Data.([]byte)
	raw := make([]byte, cmdSize+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[cmdSize:], data)
	if err := jglobal.AesEncrypt(ses.aesKey, &raw); err != nil {
		return err
	}
	size := len(raw)
	for pos := 0; pos < size; {
		n, err := ses.con.Write(raw)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
