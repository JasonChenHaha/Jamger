package jtcp

import (
	"jconfig"
	"jglobal"
	"jlog"
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
	sChan    chan *jglobal.Pack
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
		sChan:    make(chan *jglobal.Pack, 4),
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

func (ses *Ses) send(pack *jglobal.Pack) {
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
			pack := &jglobal.Pack{Id: ses.id}
			if err := recvAndDecodeToPack(pack, ses); err != nil {
				ses.tcp.delete(ses.id)
				return
			}
			ses.tcp.receive(pack)
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
			if err := encodeAndSendPack(pack, ses); err != nil {
				jlog.Error(err)
				ses.tcp.delete(ses.id)
			}
		}
	}
}
