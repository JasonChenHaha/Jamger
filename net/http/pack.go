package jhttp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"jglobal"
	"jpb"
	"juser"
)

// auth:
// client pack structure:
// +------------------------------------------------------------+
// |                            pack                            |
// +-------+---------+----------+------------+--------------+---+
// |       |   cmd   |   data   |   aeskey   |   checksum   |   |
// | rsa ( +---------+----------+------------+--------------+ ) |
// |       |    2    |   ...    |     16     |      4       |   |
// +-------+---------+----------+------------+--------------+---+

// normal:
// client pack structure:
// +---------------------------------------------------------+
// |                          pack                           |
// +---------+-------+---------+----------+--------------+---+
// |   uid   |       |   cmd   |   data   |   checksum   |   |
// +---------+ aes ( +---------+----------+--------------+ ) |
// |    4    |       |    2    |   ...    |      4       |   |
// +---------+-------+---------+----------+--------------+---+

// server pack structure:
//       +--------------------+
//       |        pack        |
//       +---------+----------+
// aes ( |   cmd   |   data   | )
//       +---------+----------+
//       |    2    |   ...    |
//       +---------+----------+

const (
	uidSize      = 4
	cmdSize      = 2
	aesKeySize   = 16
	checksumSize = 4
)

// ------------------------- package -------------------------

func decodeRSAToPack(pack *jglobal.Pack, raw []byte) error {
	if err := jglobal.RSADecrypt(jglobal.RSA_PRIVATE_KEY, &raw); err != nil {
		return err
	}
	pos := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
		return fmt.Errorf("checksum failed")
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
	pack.Data = raw[cmdSize : pos-aesKeySize]
	pack.AesKey = raw[pos-aesKeySize : pos]
	return nil
}

func decodeAESToPack(pack *jglobal.Pack, raw []byte) error {
	pack.Uid = binary.LittleEndian.Uint32(raw)
	user := juser.GetUser(pack.Uid)
	if user == nil {
		return fmt.Errorf("auth failed, uid = %d", pack.Uid)
	}
	pack.AesKey = user.AesKey
	raw = raw[uidSize:]
	if err := jglobal.AESDecrypt(pack.AesKey, &raw); err != nil {
		return err
	}
	pos := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
		return fmt.Errorf("checksum failed")
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
	pack.Data = raw[cmdSize:pos]
	return nil
}

func encodePack(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, cmdSize+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[cmdSize:], data)
	if pack.AesKey != nil {
		if err := jglobal.AESEncrypt(pack.AesKey, &raw); err != nil {
			return err
		}
	}
	pack.Data = raw
	return nil
}
