package jkcp

import (
	jconfig "jamger/config"
	jglobal "jamger/global"
	jlog "jamger/log"
	"time"

	"github.com/xtaci/kcp-go"
)

type Ses struct {
	kcp      *Kcp
	con      *kcp.UDPSession
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(kcp *Kcp, con *kcp.UDPSession, id uint64) *Ses {
	ses := &Ses{
		kcp:      kcp,
		con:      con,
		id:       id,
		rTimeout: time.Duration(jconfig.GetInt("kcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("kcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *Pack, jconfig.GetInt("kcp.sBufferSize")),
		qChan:    make(chan any, 4),
	}
	if jconfig.Get("kcp.noDelay") != nil {
		ses.con.SetNoDelay(1, jconfig.GetInt("kcp.noDelay.interval"), jconfig.GetInt("kcp.noDelay.resend"), jconfig.GetInt("kcp.noDelay.nc"))
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
			pack, err := recvPack(ses.con)
			if err != nil {
				ses.kcp.delete(ses.id)
				return
			} else {
				switch pack.Cmd {
				case jglobal.CMD_HEARTBEAT:
				case jglobal.CMD_CLOSE:
					ses.kcp.delete(ses.id)
					return
				case jglobal.CMD_PING:
					ses.kcp.Send(ses.id, jglobal.CMD_PONG, []byte{})
				default:
					ses.kcp.receive(ses.id, pack)
				}
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
			if err := sendPack(ses.con, pack); err != nil {
				jlog.Error(err)
				ses.kcp.delete(ses.id)
				return
			}
		}
	}
}
