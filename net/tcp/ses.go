package jtcp

import (
	"bytes"
	"jconfig"
	"jlog"
	"jpb"
	"net"
	"time"
)

type Ses struct {
	tcp      *Tcp
	con      *net.TCPConn
	id       uint64
	aesKey   []byte
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(tcp *Tcp, con net.Conn, id uint64) *Ses {
	ses := &Ses{
		tcp:      tcp,
		con:      con.(*net.TCPConn),
		id:       id,
		rTimeout: time.Duration(jconfig.GetInt("tcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("tcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *Pack, jconfig.GetInt("tcp.sBufferSize")),
		qChan:    make(chan any, 2),
	}
	if jconfig.GetBool("noDelay") {
		ses.con.SetNoDelay(true)
	}
	return ses
}

func (ses *Ses) run() {
	go ses.recvGoro()
	go ses.sendGoro()
}

func (ses *Ses) send(pack *Pack) {
	ses.sChan <- pack
}

func (ses *Ses) close() {
	ses.qChan <- 0
	ses.qChan <- 0
}

// ------------------------- inside -------------------------

func (ses *Ses) recvGoro() {
	for {
		select {
		case <-ses.qChan:
			return
		default:
			if ses.rTimeout > 0 {
				ses.con.SetReadDeadline(time.Now().Add(ses.rTimeout))
			}
			pack, err := recvPack(ses)
			if err != nil {
				ses.tcp.delete(ses.id)
				return
			}
			switch pack.Cmd {
			case jpb.CMD_HEARTBEAT:
			case jpb.CMD_PING:
				ses.tcp.Send(ses.id, jpb.CMD_PONG, nil)
			case jpb.CMD_SIGN_UP_REQ, jpb.CMD_SIGN_IN_REQ:
				aesKey, err := parseRSAPack(pack)
				if err != nil || bytes.Equal(aesKey, ses.aesKey) {
					ses.tcp.delete(ses.id)
					return
				}
				ses.aesKey = aesKey
				ses.tcp.receive(ses.id, pack)
			default:
				if err := parseAESpack(ses.aesKey, pack); err != nil {
					ses.tcp.delete(ses.id)
					return
				}
				ses.tcp.receive(ses.id, pack)
			}
		}
	}
}

func (ses *Ses) sendGoro() {
	defer func() {
		err := ses.con.Close()
		if err != nil {
			jlog.Error(err)
		}
	}()
	for {
		select {
		case <-ses.qChan:
			return
		case pack := <-ses.sChan:
			if ses.sTimeout > 0 {
				ses.con.SetWriteDeadline(time.Now().Add(ses.sTimeout))
			}
			if err := sendPack(ses, pack); err != nil {
				jlog.Error(err)
				ses.tcp.delete(ses.id)
				return
			}
		}
	}
}
