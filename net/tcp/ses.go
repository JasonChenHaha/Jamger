package jtcp

import (
	"io"
	jconfig "jamger/config"
	jlog "jamger/log"
	"net"
	"time"
)

type Ses struct {
	tcp      *Tcp
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSession(tcp *Tcp, id uint64) *Ses {
	return &Ses{
		tcp:      tcp,
		id:       id,
		rTimeout: time.Duration(jconfig.Get("tcp.rTimeout").(int)) * time.Second,
		sTimeout: time.Duration(jconfig.Get("tcp.sTimeout").(int)) * time.Second,
		sChan:    make(chan Pack, 666),
		qChan:    make(chan any, 4),
	}
}

func (ses *Ses) run(con net.Conn) {
	go ses.recvGoro(con)
	go ses.sendGoro(con)
}

func (ses *Ses) send(pack Pack) {
	ses.sChan <- pack
}

func (ses *Ses) close() {
	ses.qChan <- 0
	ses.qChan <- 0
}

// ------------------------- inside -------------------------

func (ses *Ses) recvGoro(con net.Conn) {
	for {
		select {
		case <-ses.qChan:
			return
		default:
			if ses.rTimeout > 0 {
				con.SetReadDeadline(time.Now().Add(ses.rTimeout))
			}
			pack, err := recvPack(con)
			if err != nil {
				if err != io.EOF {
					jlog.Error("ses recv error: ", err)
				}
				ses.tcp.delete(ses.id)
				return
			}
			ses.tcp.receive(ses.id, pack)
		}
	}
}

func (ses *Ses) sendGoro(con net.Conn) {
	defer con.Close()
	for {
		select {
		case <-ses.qChan:
			return
		case pack := <-ses.sChan:
			if ses.sTimeout > 0 {
				con.SetWriteDeadline(time.Now().Add(ses.sTimeout))
			}
			err := sendPack(con, pack)
			if err != nil {
				jlog.Error("ses send error: ", err)
				ses.tcp.delete(ses.id)
				return
			}
		}
	}
}
