package jhttp2

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"jglobal"
	"jlog"
	"jnet"
	"jpb"
	"net/http"

	"google.golang.org/protobuf/proto"
)

// ------------------------- outside -------------------------

func Init() {
	jnet.Http.RegisterPattern("/auth", authReceive)
	jnet.Http.RegisterPattern("/.well-known/pki-validation/13C96B8816330275DFED643DB3C77F41.txt", sshVerification)
}

// ------------------------- inside -------------------------

// auth client http pack structure:
// +------------------------------------------------------------+
// |                            pack                            |
// +---------+-------+----------+------------+--------------+---+
// |   cmd   |       |   data   |   aeskey   |   checksum   |   |
// |---------+ rsa ( +----------+------------+--------------+ ) |
// |    2    |       |   ...    |     16     |      4       |   |
// +---------+-------+----------+------------+--------------+---+
// server http pack structure:
// +--------------------------------+
// |              pack              |
// +-------+---------+----------+---+
// |       |   cmd   |   data   |   |
// + aes ( +---------+----------+ ) +
// |       |    2    |   ...    |   |
// +-------+---------+----------+---+
func authReceive(w http.ResponseWriter, r *http.Request) {
	// 解包
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{
		Cmd: jpb.CMD(binary.LittleEndian.Uint16(body)),
	}
	body = body[jglobal.CMD_SIZE:]
	if err := jglobal.RsaDecrypt(jglobal.RSA_PRIVATE_KEY, &body); err != nil {
		return
	}
	pos := len(body) - jglobal.CHECKSUM_SIZE
	if binary.LittleEndian.Uint32(body[pos:]) != crc32.ChecksumIEEE(body[:pos]) {
		jlog.Error("checksum failed")
		return
	}
	pack.Data = body[:pos-jglobal.AESKEY_SIZE]
	pack.Ctx = body[pos-jglobal.AESKEY_SIZE : pos]
	// 执行
	han := jnet.Http.Handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.Template)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.Fun(pack)
	} else {
		han = jnet.Http.Handler[jpb.CMD_TRANSFER]
		if han == nil {
			jlog.Error("no cmd(TRANSFER).")
			return
		}
		han.Fun(pack)
	}
	// 打包
	if v, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(v)
		if err != nil {
			jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	data := pack.Data.([]byte)
	raw := make([]byte, jglobal.CMD_SIZE+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[jglobal.CMD_SIZE:], data)
	if pack.Ctx != nil {
		if err := jglobal.AesEncrypt(pack.Ctx.([]byte), &raw); err != nil {
			return
		}
	}
	// 发送
	if _, err = w.Write(raw); err != nil {
		jlog.Error(err)
	}
}

func sshVerification(w http.ResponseWriter, r *http.Request) {
	data := []byte("0DC65D90903F77218EF6C3A318C39CF4B6ACA169E1D7E50C9E6EE01E40849465\ncomodoca.com\nacc473308f41d9c")
	if _, err := w.Write(data); err != nil {
		jlog.Error(err)
	}
}
