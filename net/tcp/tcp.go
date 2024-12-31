package jtcp

import (
	"crypto/rsa"
	"jconfig"
	"jlog"
	"jtrash"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
)

type Callback func(id uint64, cmd uint16, msg proto.Message)

type Handler struct {
	cb  Callback
	obj proto.Message
}

type Tcp struct {
	idc        uint64
	ses        sync.Map
	counter    uint64
	handler    map[uint16]*Handler
	privateKey *rsa.PrivateKey
}

// ------------------------- outside -------------------------

func NewTcp() *Tcp {
	key, err := jtrash.RSALoadPrivateKey(jconfig.GetString("rsa.privateKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	tcp := &Tcp{
		handler:    make(map[uint16]*Handler),
		privateKey: key,
	}
	addr := jconfig.GetString("tcp.addr")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info("listen on ", addr)
	go tcp.accept(listener)
	if jconfig.GetBool("debug") {
		go tcp.watch()
	}
	return tcp
}

func (tcp *Tcp) Register(id uint16, cb Callback, obj proto.Message) {
	tcp.handler[id] = &Handler{
		cb:  cb,
		obj: obj,
	}
}

func (tcp *Tcp) Send(id uint64, cmd uint16, data []byte) {
	obj, ok := tcp.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Ses).send(makePack(cmd, data))
}

// ------------------------- package -------------------------

func (tcp *Tcp) receive(id uint64, pack *Pack) {
	han, ok := tcp.handler[pack.Cmd]
	if !ok {
		jlog.Warn("cmd not exist, ", pack.Cmd)
		tcp.delete(id)
		return
	}
	data, err := jtrash.RSADecrypt(tcp.privateKey, pack.Data)
	if err != nil {
		jlog.Warnf("%s, %d", err, pack.Cmd)
		tcp.delete(id)
		return
	}
	err = proto.Unmarshal(data, han.obj)
	if err != nil {
		jlog.Warnf("%s, %d", err, pack.Cmd)
		tcp.delete(id)
		return
	}
	han.cb(id, pack.Cmd, han.obj)
}

// ------------------------- inside -------------------------

func (tcp *Tcp) accept(listener net.Listener) {
	for {
		con, err := listener.Accept()
		if err != nil {
			jlog.Error(err)
			continue
		} else {
			tcp.add(con)
		}
	}
}

func (tcp *Tcp) add(con net.Conn) {
	id := atomic.AddUint64(&tcp.idc, 1)
	ses := newSes(tcp, con, id)
	tcp.ses.Store(id, ses)
	tcp.counter++
	ses.run()
}

func (tcp *Tcp) delete(id uint64) {
	if obj, ok := tcp.ses.Load(id); ok {
		tcp.ses.Delete(id)
		tcp.counter--
		obj.(*Ses).close()
	}
}

// ------------------------- debug -------------------------

func (tcp *Tcp) watch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		jlog.Debug("connecting ", tcp.counter)
	}
}
