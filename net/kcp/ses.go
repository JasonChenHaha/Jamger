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
	o := &Ses{
		kc:       kc,
		con:      con,
		id:       id,
		rTimeout: time.Duration(jconfig.GetInt("kcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("kcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *Pack, 4),
		qChan:    make(chan any, 4),
	}
	if jconfig.Get("kcp.noDelay") != nil {
		o.con.SetNoDelay(1, jconfig.GetInt("kcp.noDelay.interval"), jconfig.GetInt("kcp.noDelay.resend"), jconfig.GetInt("kcp.noDelay.nc"))
	}
	return o
}

func (o *Ses) run() {
	go o.recvGoro()
	go o.sendGoro()
}

func (o *Ses) send(pack *Pack) {
	o.sChan <- pack
}

func (o *Ses) close() {
	o.qChan <- 0
	o.qChan <- 0
}

// ------------------------- inside -------------------------

func (o *Ses) recvGoro() {
	for {
		select {
		case <-o.qChan:
			return
		default:
			if o.rTimeout > 0 {
				o.con.SetReadDeadline(time.Now().Add(o.rTimeout))
			}
			pack, err := recvPack(o.con)
			if err != nil {
				o.kc.delete(o.id)
				return
			} else {
				switch pack.Cmd {
				case jpb.CMD_HEARTBEAT:
				case jpb.CMD_CLOSE:
					o.kc.delete(o.id)
					return
				default:
					o.kc.receive(o.id, pack)
				}
			}
		}
	}
}

func (o *Ses) sendGoro() {
	defer func() {
		err := o.con.Close()
		if err != nil {
			jlog.Error(err)
		}
	}()
	for {
		select {
		case <-o.qChan:
			return
		case pack := <-o.sChan:
			if o.sTimeout > 0 {
				o.con.SetWriteDeadline(time.Now().Add(o.sTimeout))
			}
			if err := sendPack(o.con, pack); err != nil {
				jlog.Error(err)
				o.kc.delete(o.id)
				return
			}
		}
	}
}
