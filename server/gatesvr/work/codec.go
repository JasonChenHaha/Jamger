package jwork

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"jglobal"
	"jpb"
	"juser"
)

const (
	uidSize      = 4
	cmdSize      = 2
	aesKeySize   = 16
	checksumSize = 4
)

// ------------------------- package -------------------------

// server tcp pack structure:
// +-------------------------------------------+
// |                   pack                    |
// +----------+-------+---------+----------+---+
// |   size   |       |   cmd   |   data   |   |
// +----------+ aes ( +---------+----------+ ) |
// |    2     |       |    2    |   size   |   |
// +----------+-------+---------+----------+---+

func tcpEncode(pack *jglobal.Pack) error {
	user := pack.User.(*juser.User)
	user.Lock()
	data := pack.Data.([]byte)
	raw := make([]byte, cmdSize+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[cmdSize:], data)
	if err := jglobal.AesEncrypt(user.AesKey, &raw); err != nil {
		return err
	}
	pack.Data = raw
	return nil
}

// client tcp pack structure:
// +--------------------------------------------------------------------+
// |                               pack                                 |
// +----------+---------+-------+---------+----------+--------------+---+
// |   size   |   uid   |       |   cmd   |   data   |   checksum   |   |
// +----------+---------+ aes ( +---------+----------+--------------+ ) |
// |    2     |    4    |       |    2    |   ...    |      4       |   |
// +----------+---------+-------+---------+----------+--------------+---+

func tcpDecode(pack *jglobal.Pack) (uint32, error) {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	user := juser.GetUser(uid)
	if user == nil {
		return 0, fmt.Errorf("no such user, uid = %d", uid)
	}
	user.Lock()
	pack.User = user
	raw = raw[uidSize:]
	if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
		return 0, err
	}
	posChecksum := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[posChecksum:]) != crc32.ChecksumIEEE(raw[:posChecksum]) {
		return 0, fmt.Errorf("checksum failed")
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[:cmdSize]))
	pack.Data = raw[cmdSize:posChecksum]
	return uid, nil
}

// server http pack structure:
//       +--------------------+
//       |        pack        |
//       +---------+----------+
// aes ( |   cmd   |   data   | )
//       +---------+----------+
//       |    2    |   ...    |
//       +---------+----------+

func httpEncode(url string, pack *jglobal.Pack) error {
	if url == "/" {
		user := pack.User.(*juser.User)
		user.Unlock()
		data := pack.Data.([]byte)
		raw := make([]byte, cmdSize+len(data))
		binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
		copy(raw[cmdSize:], data)
		if user.AesKey != nil {
			if err := jglobal.AesEncrypt(user.AesKey, &raw); err != nil {
				return err
			}
		}
		pack.Data = raw
	} else {
		data := pack.Data.([]byte)
		raw := make([]byte, cmdSize+len(data))
		binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
		copy(raw[cmdSize:], data)
		if pack.User != nil {
			if err := jglobal.AesEncrypt(pack.User.([]byte), &raw); err != nil {
				return err
			}
		}
		pack.Data = raw
	}
	return nil
}

// auth client http pack structure:
// +------------------------------------------------------------+
// |                            pack                            |
// +-------+---------+----------+------------+--------------+---+
// |       |   cmd   |   data   |   aeskey   |   checksum   |   |
// | rsa ( +---------+----------+------------+--------------+ ) |
// |       |    2    |   ...    |     16     |      4       |   |
// +-------+---------+----------+------------+--------------+---+
// client http pack structure:
// +---------------------------------------------------------+
// |                          pack                           |
// +---------+-------+---------+----------+--------------+---+
// |   uid   |       |   cmd   |   data   |   checksum   |   |
// +---------+ aes ( +---------+----------+--------------+ ) |
// |    4    |       |    2    |   ...    |      4       |   |
// +---------+-------+---------+----------+--------------+---+

func httpDecode(url string, pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	if url == "/" {
		uid := binary.LittleEndian.Uint32(raw)
		user := juser.GetUser(uid)
		if user == nil {
			return fmt.Errorf("no such user, uid = %d", uid)
		}
		user.Lock()
		pack.User = user
		raw = raw[uidSize:]
		if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
			return err
		}
		pos := len(raw) - checksumSize
		if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
			return fmt.Errorf("checksum failed")
		}
		pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
		pack.Data = raw[cmdSize:pos]
	} else {
		if err := jglobal.RsaDecrypt(jglobal.RSA_PRIVATE_KEY, &raw); err != nil {
			return err
		}
		pos := len(raw) - checksumSize
		if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
			return fmt.Errorf("checksum failed")
		}
		pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
		pack.Data = raw[cmdSize : pos-aesKeySize]
		pack.User = raw[pos-aesKeySize : pos] // 取巧设计
	}
	return nil
}

// rpc pack structure:
// +------------------------------+
// |             pack             |
// +---------+---------+----------+
// |   uid   |   cmd   |   data   |
// +---------+---------+----------+
// |    4    |    2    |   ...    |
// +------------------------------+

func rpcEncode(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, uidSize+cmdSize+len(data))
	if user, ok := pack.User.(*juser.User); ok {
		binary.LittleEndian.PutUint32(raw, uint32(user.Uid))
	}
	binary.LittleEndian.PutUint16(raw[uidSize:], uint16(pack.Cmd))
	copy(raw[uidSize+cmdSize:], data)
	pack.Data = raw
	return nil
}

func rpcDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	if pack.User == nil && uid != 0 {
		user := juser.GetUser(uid)
		if user == nil {
			return fmt.Errorf("no such user, uid = %d", uid)
		}
		user.Lock()
		pack.User = user
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize:]))
	pack.Data = raw[uidSize+cmdSize:]
	return nil
}
