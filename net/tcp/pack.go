package jtcp

import (
	"jglobal"
	"jpb"
)

// pack structure:
// |                                pack                                 |
// +---------+---------+---------+--------------+----------+-------------|
// |   cmd   |   cmd   |   uid   |   (aeskey)   |   size   |   payload   |
// +---------+---------+---------+--------------+----------+-------------|
// |    2    |    2    |    4    |      4       |    2     |    size     |

// sign up or sign in:
// client: cmd + rsa(cmd+uid+aeskey+size+payload) -> server
// server: cmd + aes(payload) -> client

// other:
// client: cmd + aes(cmd+uid+size+payload) -> server
// server: aes(payload) -> client

const (
	CmdSize    = 2
	UidSize    = 4
	AesKeySize = 16
	SizeSize   = 2
)

type Pack struct {
	Cmd  jpb.CMD
	Data []byte
}

// ------------------------- package -------------------------

func recvPack(ses *Ses) (pack *Pack, err error) {
	// 读数据
	// buffer := make([]byte, HeadSize)
	// if _, err = io.ReadFull(ses.con, buffer); err != nil {
	// 	return
	// }
	// bodySize := binary.LittleEndian.Uint16(buffer)
	// buffer = make([]byte, bodySize)
	// if _, err = io.ReadFull(ses.con, buffer); err != nil {
	// 	return
	// }
	// pack = &Pack{
	// 	Cmd:  jpb.CMD(binary.LittleEndian.Uint16(buffer)),
	// 	Data: buffer[CmdSize:],
	// }
	return
}

func sendPack(ses *Ses, pack *Pack) error {
	// bodySize := CmdSize + len(pack.Data)
	// size := HeadSize + bodySize
	// buffer := make([]byte, size)
	// binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	// binary.LittleEndian.PutUint16(buffer[HeadSize:], uint16(pack.Cmd))
	// copy(buffer[HeadSize+CmdSize:], pack.Data)
	// // 写数据
	// for pos := 0; pos < size; {
	// 	n, err := ses.con.Write(buffer)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	pos += n
	// }
	return nil
}

func parseRSAPack(pack *Pack) ([]byte, error) {
	if err := jglobal.RSADecrypt(jglobal.RSA_PRIVATE_KEY, &pack.Data); err != nil {
		return nil, err
	}
	size := len(pack.Data) - AesKeySize
	aesKey := pack.Data[size:]
	pack.Data = pack.Data[:size]
	return aesKey, nil
}
