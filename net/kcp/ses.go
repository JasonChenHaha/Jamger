package jkcp

import (
	"jconfig"
	"jlog"
	"jpb"
	"time"

	"github.com/xtaci/kcp-go"
)

type Ses struct {
	kc       *Kcp
	con      *kcp.UDPSession
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(kc *Kcp, con *kcp.UDPSession, id uint64) *Ses {
	ses := &Ses{
		kc:       kc,
		con:      con,
		id:       id,
		rTimeout: time.Duration(jconfig.GetInt("kcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("kcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *Pack, 4),
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
				ses.kc.delete(ses.id)
				return
			} else {
				switch pack.Cmd {
				case jpb.CMD_HEARTBEAT:
				case jpb.CMD_CLOSE:
					ses.kc.delete(ses.id)
					return
				default:
					ses.kc.receive(ses.id, pack)
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
				ses.kc.delete(ses.id)
				return
			}
		}
	}
}
