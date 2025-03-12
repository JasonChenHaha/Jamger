package jwork

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"jglobal"
	"jlog"
	"jpb"
	"juser"
)

const (
	uidSize      = 4
	gateSize     = 4
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
	user := pack.Ctx.(*juser.User)
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
// +----------+---------+---------+-------+----------+--------------+---+
// |   size   |   uid   |   cmd   |       |   data   |   checksum   |   |
// +----------+---------+---------+ aes ( +----------+--------------+ ) |
// |    2     |    4    |    2    |       |   ...    |      4       |   |
// +----------+---------+---------+-------+----------+--------------+---+

func tcpDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize:]))
	var user *juser.User
	if pack.Cmd == jpb.CMD_LOGIN_REQ {
		if user = juser.GetUser(uid); user == nil {
			user = juser.NewUser(uid)
		}
		user.Load()
	} else {
		if user = juser.GetUser(uid); user == nil {
			err := fmt.Errorf("no such user, uid(%d)", uid)
			jlog.Error(err)
			return err
		}
	}
	pack.Ctx = user
	raw = raw[uidSize+cmdSize:]
	if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
		jlog.Error(err)
		return err
	}
	pos := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
		err := fmt.Errorf("checksum failed")
		jlog.Error(err)
		return err
	}
	pack.Data = raw[:pos]
	return nil
}

// server http pack structure:
// +--------------------------------+
// |              pack              |
// +-------+---------+----------+---+
// |       |   cmd   |   data   |   |
// + aes ( +---------+----------+ ) +
// |       |    2    |   ...    |   |
// +-------+---------+----------+---+

func httpEncode(url string, pack *jglobal.Pack) error {
	if url == "/" {
		user := pack.Ctx.(*juser.User)
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
		if pack.Ctx != nil {
			if err := jglobal.AesEncrypt(pack.Ctx.([]byte), &raw); err != nil {
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
// +---------+-------+----------+------------+--------------+---+
// |   cmd   |       |   data   |   aeskey   |   checksum   |   |
// |---------+ rsa ( +----------+------------+--------------+ ) |
// |    2    |       |   ...    |     16     |      4       |   |
// +---------+-------+----------+------------+--------------+---+
// client http pack structure:
// +---------------------------------------------------------+
// |                          pack                           |
// +---------+---------+-------+----------+--------------+---+
// |   uid   |   cmd   |       |   data   |   checksum   |   |
// +---------+---------+ aes ( +----------+--------------+ ) |
// |    4    |    2    |       |   ...    |      4       |   |
// +---------+---------+-------+----------+--------------+---+

func httpDecode(url string, pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	if url == "/" {
		uid := binary.LittleEndian.Uint32(raw)
		// to do: 每次请求都会重新load
		user := juser.GetUser(uid)
		if user == nil {
			user = juser.NewUser(uid)
		}
		user.Load()
		pack.Ctx = user
		pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize:]))
		raw = raw[uidSize+cmdSize:]
		if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
			return err
		}
		pos := len(raw) - checksumSize
		if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
			err := fmt.Errorf("checksum failed")
			jlog.Error(err)
			return err
		}
		pack.Data = raw[:pos]
	} else {
		pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
		raw = raw[cmdSize:]
		if err := jglobal.RsaDecrypt(jglobal.RSA_PRIVATE_KEY, &raw); err != nil {
			return err
		}
		pos := len(raw) - checksumSize
		if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
			err := fmt.Errorf("checksum failed")
			jlog.Error(err)
			return err
		}
		pack.Data = raw[:pos-aesKeySize]
		pack.Ctx = raw[pos-aesKeySize : pos]
	}
	return nil
}

// client websocket pack structure:
// +--------------------------------------+
// |                 pack                 |
// +---------+---------+-------+----------+
// |   uid   |   cmd   |       |   data   |
// +---------+---------+ aes ( +----------+
// |    4    |   2     |       |    ..    |
// +---------+---------+-------+----------+
func webDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize:]))
	var user *juser.User
	if pack.Cmd == jpb.CMD_LOGIN_REQ {
		if user = juser.GetUser(uid); user == nil {
			user = juser.NewUser(uid)
		}
		user.Load()
	} else {
		if user = juser.GetUser(uid); user == nil {
			err := fmt.Errorf("no such user, uid(%d)", uid)
			jlog.Error(err)
			return err
		}
	}
	pack.Ctx = user
	raw = raw[uidSize+cmdSize:]
	if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
		return err
	}
	pos := len(raw) - checksumSize
	if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
		err := fmt.Errorf("checksum failed")
		jlog.Error(err)
		return err
	}
	pack.Data = raw[:pos]
	return nil
}

// server websocket pack structure:
// +--------------------------------+
// |              pack              |
// +-------+---------+----------+---+
// |       |   cmd   |   data   |   |
// + aes ( +---------+----------+ ) +
// |       |    2    |   ...    |   |
// +-------+---------+----------+---+
func webEncode(pack *jglobal.Pack) error {
	user := pack.Ctx.(*juser.User)
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

// rpc pack head structure:
// +------------------------------+
// |             head             |
// +---------+----------+---------+
// |   uid   |   gate   |   cmd   |
// +---------+----------+---------+
// |    4    |    4     |    2    |
// +------------------------------+
//   \         /
//    \      /
//     |   /
//     | /
// rpc pack structure:
// +---------------------+
// |        pack         |
// +----------+----------+
// |   head   |   data   |
// +----------+----------+
// |    ..    |    ..    |
// +----------+----------+

func rpcEncode(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, uidSize+gateSize+cmdSize+len(data))
	if user, ok := pack.Ctx.(*juser.User); ok {
		binary.LittleEndian.PutUint32(raw, user.Uid)
		binary.LittleEndian.PutUint32(raw[uidSize:], uint32(user.Gate))
	}
	binary.LittleEndian.PutUint16(raw[uidSize+gateSize:], uint16(pack.Cmd))
	copy(raw[uidSize+gateSize+cmdSize:], data)
	pack.Data = raw
	return nil
}

func rpcDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	if pack.Ctx == nil && uid != 0 {
		user := juser.GetUser(uid)
		if user != nil {
			pack.Ctx = user
		} else {
			pack.Ctx = uid
		}
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[uidSize+gateSize:]))
	pack.Data = raw[uidSize+gateSize+cmdSize:]
	return nil
}
