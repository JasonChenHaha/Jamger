package jtcp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"jglobal"
	"jpb"
)

// client pack structure:
// +--------------------------------------------------------------------+
// |                               pack|                                |
// +----------+-------+---------+---------+----------+--------------+---+
// |   size   |       |   uid   |   cmd   |   data   |   checksum   |   |
// +----------+ aes ( +---------+---------+----------+--------------+ ) |
// |    2     |       |    4    |    2    |   ...    |      4       |   |
// +----------+-------+---------+---------+----------+--------------+---+

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
	if err := jglobal.RSADecrypt(jglobal.RSA_PRIVATE_KEY, &raw); err != nil {
		return err
	}
	posChecksum := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[posChecksum:]) != crc32.ChecksumIEEE(raw[:posChecksum]) {
		return fmt.Errorf("checksum failed")
	}
	pack.Uid = binary.LittleEndian.Uint32(raw)
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize:]))
	pack.Data = raw[uidSize+cmdSize : posChecksum]
	return nil
}

// 加密、发送数据
func encodeAndSendPack(pack *jglobal.Pack, ses *Ses) error {
	data := pack.Data.([]byte)
	if err := jglobal.AESEncrypt(ses.aesKey, &data); err != nil {
		return err
	}
	size := cmdSize + len(data)
	raw := make([]byte, size)
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[cmdSize:], data)
	for pos := 0; pos < size; {
		n, err := ses.con.Write(raw)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
