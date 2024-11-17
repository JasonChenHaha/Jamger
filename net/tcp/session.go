package tcp

import (
	jconfig "jamger/config"
	jlog "jamger/log"
	"net"
	"time"
)

var g_callback func(id uint64, pack Pack)

type Session struct {
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan Pack
	qChan    chan any
}

// ------------------------- outside -------------------------

func SetCallback(f func(id uint64, pack Pack)) {
	g_callback = f
}

// ------------------------- package -------------------------

func newSession(id uint64) *Session {
	return &Session{
		id:       id,
		rTimeout: jconfig.Get("tcp.rTimeout").(time.Duration),
		sTimeout: jconfig.Get("tcp.sTimeout").(time.Duration),
		sChan:    make(chan Pack, 666),
		qChan:    make(chan any, 4),
	}
}

func (ses *Session) run(con net.Conn) {
	go ses.recvGoro(con)
	go ses.sendGoro(con)
}

func (ses *Session) send(pack Pack) {
	ses.sChan <- pack
}

func (ses *Session) close() {
	ses.qChan <- 0
	ses.qChan <- 0
}

// ------------------------- inside -------------------------

func (ses *Session) recvGoro(con net.Conn) {
	select {
	case <-ses.qChan:
		break
	default:
		con.SetReadDeadline(time.Now().Add(ses.rTimeout))
		pack, err := recvPack(con)
		if err != nil {
			jlog.Errorln("session recv error: ", err)
			g_sesMgr.close(ses.id)
		}
		g_callback(ses.id, pack)
	}
}

func (ses *Session) sendGoro(con net.Conn) {
	select {
	case <-ses.qChan:
		break
	case pack := <-ses.sChan:
		con.SetWriteDeadline(time.Now().Add(ses.sTimeout))
		err := sendPack(con, pack)
		if err != nil {
			jlog.Errorln("session send error: ", err)
			g_sesMgr.close(ses.id)
		}
	}
	con.Close()
}
