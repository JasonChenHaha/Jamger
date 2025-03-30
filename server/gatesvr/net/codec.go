package jnet2

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"jglobal"
	"jlog"
	"jpb"
	"juser2"
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
	user := pack.Ctx.(*juser2.User)
	data := pack.Data.([]byte)
	raw := make([]byte, jglobal.CMD_SIZE+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[jglobal.CMD_SIZE:], data)
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
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[jglobal.UID_SIZE:]))
	var user *juser2.User
	if pack.Cmd == jpb.CMD_LOGIN_REQ {
		if user = juser2.GetUser(uid); user == nil {
			user = juser2.NewUser(uid)
		}
		user.Load()
	} else {
		if user = juser2.GetUser(uid); user == nil {
			err := fmt.Errorf("no such user, uid(%d)", uid)
			jlog.Error(err)
			return err
		}
	}
	pack.Ctx = user
	raw = raw[jglobal.UID_SIZE+jglobal.CMD_SIZE:]
	if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
		jlog.Error(err)
		return err
	}
	pos := len(raw) - jglobal.CHECKSUM_SIZE
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

func httpEncode(pack *jglobal.Pack) error {
	user := pack.Ctx.(*juser2.User)
	data := pack.Data.([]byte)
	raw := make([]byte, jglobal.CMD_SIZE+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[jglobal.CMD_SIZE:], data)
	if user.AesKey != nil {
		if err := jglobal.AesEncrypt(user.AesKey, &raw); err != nil {
			return err
		}
	}
	pack.Data = raw
	return nil
}

// client http pack structure:
// +---------------------------------------------------------+
// |                          pack                           |
// +---------+---------+-------+----------+--------------+---+
// |   uid   |   cmd   |       |   data   |   checksum   |   |
// +---------+---------+ aes ( +----------+--------------+ ) |
// |    4    |    2    |       |   ...    |      4       |   |
// +---------+---------+-------+----------+--------------+---+
func httpDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	user := juser2.GetUser(uid)
	if user == nil {
		user = juser2.NewUser(uid)
	}
	// to do: 每次请求都会重新load
	user.Load()
	pack.Ctx = user
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[jglobal.UID_SIZE:]))
	raw = raw[jglobal.UID_SIZE+jglobal.CMD_SIZE:]
	if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
		return err
	}
	pos := len(raw) - jglobal.CHECKSUM_SIZE
	if binary.LittleEndian.Uint32(raw[pos:]) != crc32.ChecksumIEEE(raw[:pos]) {
		err := fmt.Errorf("checksum failed")
		jlog.Error(err)
		return err
	}
	pack.Data = raw[:pos]
	return nil
}

// server https pack structure:
// +--------------------+
// |        pack        |
// +---------+----------+
// |   cmd   |   data   |
// |---------+----------+
// |    2    |   ...    |
// +---------+----------+

func httpsEncode(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	raw := make([]byte, jglobal.CMD_SIZE+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[jglobal.CMD_SIZE:], data)
	pack.Data = raw
	return nil
}

// client https pack structure:
// +------------------------------+
// |             pack             |
// +---------+---------+----------+
// |   uid   |   cmd   |   data   |
// +---------+---------+----------+
// |    4    |    2    |   ...    |
// +---------+---------+----------+

func httpsDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	user := juser2.GetUser(uid)
	if user == nil {
		user = juser2.NewUser(uid)
	}
	// to do: 每次请求都会重新load
	// user.Load()
	pack.Ctx = user
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[jglobal.UID_SIZE:]))
	pack.Data = raw[jglobal.UID_SIZE+jglobal.CMD_SIZE:]
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
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[jglobal.UID_SIZE:]))
	var user *juser2.User
	if pack.Cmd == jpb.CMD_LOGIN_REQ {
		if user = juser2.GetUser(uid); user == nil {
			user = juser2.NewUser(uid)
		}
		user.Load()
	} else {
		if user = juser2.GetUser(uid); user == nil {
			err := fmt.Errorf("no such user, uid(%d)", uid)
			jlog.Error(err)
			return err
		}
	}
	pack.Ctx = user
	raw = raw[jglobal.UID_SIZE+jglobal.CMD_SIZE:]
	if err := jglobal.AesDecrypt(user.AesKey, &raw); err != nil {
		return err
	}
	pos := len(raw) - jglobal.CHECKSUM_SIZE
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
	user := pack.Ctx.(*juser2.User)
	data := pack.Data.([]byte)
	raw := make([]byte, jglobal.CMD_SIZE+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[jglobal.CMD_SIZE:], data)
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
	raw := make([]byte, jglobal.UID_SIZE+jglobal.GATE_SIZE+jglobal.CMD_SIZE+len(data))
	if user, ok := pack.Ctx.(*juser2.User); ok {
		binary.LittleEndian.PutUint32(raw, user.Uid)
		binary.LittleEndian.PutUint32(raw[jglobal.UID_SIZE:], uint32(user.Gate))
	}
	binary.LittleEndian.PutUint16(raw[jglobal.UID_SIZE+jglobal.GATE_SIZE:], uint16(pack.Cmd))
	copy(raw[jglobal.UID_SIZE+jglobal.GATE_SIZE+jglobal.CMD_SIZE:], data)
	pack.Data = raw
	return nil
}

func rpcDecode(pack *jglobal.Pack) error {
	raw := pack.Data.([]byte)
	uid := binary.LittleEndian.Uint32(raw)
	if pack.Ctx == nil && uid != 0 {
		user := juser2.GetUser(uid)
		if user != nil {
			pack.Ctx = user
		} else {
			pack.Ctx = uid
		}
	}
	pack.Cmd = jpb.CMD(binary.LittleEndian.Uint16(raw[jglobal.UID_SIZE+jglobal.GATE_SIZE:]))
	pack.Data = raw[jglobal.UID_SIZE+jglobal.GATE_SIZE+jglobal.CMD_SIZE:]
	return nil
}
