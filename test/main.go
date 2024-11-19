package main

import (
	"encoding/binary"
	"io"
	jconfig "jamger/config"
	jlog "jamger/log"
	jtcp "jamger/net/tcp"
	"net"
	"net/http"
	"strings"
)

func main() {
	jlog.Info("<test start>")
	testTcp()
	// testHttp()
}

func testTcp() {
	jlog.Info("<test tcp>")
	addr := strings.Split(jconfig.Get("tcp.addr").(string), ":")
	con, err := net.Dial("tcp", "127.0.0.1:"+addr[1])
	if err != nil {
		jlog.Fatal(err)
	}
	defer con.Close()
	jlog.Info("connect to server ", addr)

	pack := jtcp.Pack{
		Cmd:  2,
		Data: []byte("hello world"),
	}
	sendPack(con, pack)

	pack = recvPack(con)
	jlog.Info(pack.Cmd, string(pack.Data))
}

func testHttp() {
	jlog.Info("<test http>")
	rsp, err := http.Get("http://127.0.0.1:8080?abc=1&ddd=2&haha=3")
	if err != nil {
		jlog.Fatal(err)
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info(string(body))
}

func sendPack(con net.Conn, pack jtcp.Pack) {
	bodySize := jtcp.CMD_SIZE + len(pack.Data)
	size := jtcp.HEAD_SIZE + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[jtcp.HEAD_SIZE:], pack.Cmd)
	copy(buffer[jtcp.HEAD_SIZE+jtcp.CMD_SIZE:], pack.Data)
	for pos := 0; pos < size; {
		n, err := con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func recvPack(con net.Conn) (pack jtcp.Pack) {
	buffer := make([]byte, jtcp.HEAD_SIZE)
	_, err := io.ReadFull(con, buffer)
	if err != nil {
		jlog.Fatal(err)
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	_, err = io.ReadFull(con, buffer)
	if err != nil {
		jlog.Fatal(err)
	}
	pack.Cmd = binary.LittleEndian.Uint16(buffer)
	pack.Data = buffer[jtcp.CMD_SIZE:]
	return pack
}
