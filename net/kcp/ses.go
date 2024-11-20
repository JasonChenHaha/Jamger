package jkcp

import (
	"io"
	jconfig "jamger/config"
	jglobal "jamger/global"
	jlog "jamger/log"
	"time"

	"github.com/xtaci/kcp-go"
)

type Ses struct {
	kcp      *Kcp
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(kcp *Kcp, id uint64) *Ses {
	return &Ses{
		kcp:      kcp,
		id:       id,
		rTimeout: time.Duration(jconfig.Get("kcp.rTimeout").(int)) * time.Second,
		sTimeout: time.Duration(jconfig.Get("kcp.sTimeout").(int)) * time.Second,
		sChan:    make(chan *Pack, jglobal.G_KCP_SEND_BUFFER_SIZE),
		qChan:    make(chan any, 4),
	}
}

func (ses *Ses) run(con *kcp.UDPSession) {
	go ses.recvGoro(con)
	go ses.sendGoro(con)
}

func (ses *Ses) send(pack *Pack) {
	ses.sChan <- pack
}

func (ses *Ses) close() {
	ses.qChan <- 0
	ses.qChan <- 0
}

// ------------------------- inside -------------------------

func (ses *Ses) recvGoro(con *kcp.UDPSession) {
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
					jlog.Error(err)
				}
				ses.kcp.delete(ses.id)
				return
			}
			ses.kcp.receive(ses.id, pack)
		}
	}
}

func (ses *Ses) sendGoro(con *kcp.UDPSession) {
	defer func() {
		err := con.Close()
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
				con.SetWriteDeadline(time.Now().Add(ses.sTimeout))
			}
			if err := sendPack(con, pack); err != nil {
				jlog.Error(err)
				ses.kcp.delete(ses.id)
				return
			}
		}
	}
}
