package jhttp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"jglobal"
	"jpb"
)

// client pack structure:
//       +------------------------------------------------+
//       |                      pack                      |
//       +---------+----------+------------+--------------+
// rsa ( |   cmd   |   data   |   aeskey   |   checksum   | )
//       +---------+----------+------------+--------------+
//       |    2    |   size   |     16     |      4       |
//       +---------+----------+------------+--------------+

// server pack structure:
//       +--------------------+
//       |        pack        |
//       +---------+----------+
// aes ( |   cmd   |   data   | )
//       +---------+----------+
//       |    2    |   size   |
//       +---------+----------+

const (
	cmdSize      = 2
	aesKeySize   = 16
	checksumSize = 4
)

type Pack struct {
	cmd    jpb.CMD
	data   []byte
	aesKey []byte
}

// ------------------------- package -------------------------

func decodeToPack(raw []byte) (*Pack, error) {
	if err := jglobal.RSADecrypt(jglobal.RSA_PRIVATE_KEY, &raw); err != nil {
		return nil, err
	}
	size := len(raw)
	posAes, posChecksum := size-checksumSize-aesKeySize, size-checksumSize
	checksum := binary.LittleEndian.Uint32(raw[posChecksum:])
	if checksum != crc32.ChecksumIEEE(raw[:posChecksum]) {
		return nil, fmt.Errorf("checksum failed")
	}
	pack := &Pack{
		cmd:    jpb.CMD(binary.LittleEndian.Uint16(raw)),
		data:   raw[cmdSize:posAes],
		aesKey: raw[posAes:posChecksum],
	}
	return pack, nil
}

func encodeFromPack(pack *Pack) ([]byte, error) {
	raw := make([]byte, cmdSize+len(pack.data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.cmd))
	copy(raw[cmdSize:], pack.data)
	if err := jglobal.AESEncrypt(pack.aesKey, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
