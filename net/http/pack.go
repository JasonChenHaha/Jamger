package jhttp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"jglobal"
	"jpb"
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

func decodeRsaToPack(pack *jglobal.Pack, raw []byte) error {
	if err := jglobal.RsaDecrypt(jglobal.RSA_PRIVATE_KEY, &raw); err != nil {
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

func decodeAesToPack(pack *jglobal.Pack, raw []byte) error {
	// uid := binary.LittleEndian.Uint32(raw)
	// user := juser.GetUser(uid)
	// if user == nil {
	// 	return fmt.Errorf("auth failed, uid = %d", uid)
	// }
	// pack.AesKey = user.GetAesKey()
	// raw = raw[uidSize:]
	// if err := jglobal.AesDecrypt(pack.AesKey, &raw); err != nil {
	// 	return err
	// }
	// pos := len(raw) - checksumSize
	// if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
	// 	return fmt.Errorf("checksum failed")
	// }
	// pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
	// pack.Data = raw[cmdSize:pos]
	return nil
}

func encodePack(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, cmdSize+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[cmdSize:], data)
	if pack.AesKey != nil {
		if err := jglobal.AesEncrypt(pack.AesKey, &raw); err != nil {
			return err
		}
	}
	pack.Data = raw
	return nil
}
