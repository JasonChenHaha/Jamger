package jweb

import (
	"jconfig"
	"jlog"
	"jpb"
	"time"

	"github.com/gorilla/websocket"
)

type Ses struct {
	ws       *WebSvr
	con      *websocket.Conn
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(ws *WebSvr, con *websocket.Conn, id uint64) *Ses {
	ses := &Ses{
		ws:       ws,
		con:      con,
		id:       id,
		rTimeout: time.Duration(jconfig.GetInt("tcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("tcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *Pack, 4),
		qChan:    make(chan any, 2),
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
			_, data, err := ses.con.ReadMessage()
			if err != nil {
				ses.ws.delete(ses.id)
				return
			}
			pack := unserializeData(data)
			switch pack.Cmd {
			case jpb.CMD_HEARTBEAT:
			case jpb.CMD_PING:
				ses.ws.Send(ses.id, jpb.CMD_PONG, []byte{})
			default:
				ses.ws.receive(ses.id, pack)
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
			data := serializePack(pack)
			err := ses.con.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				jlog.Error(err)
				return
			}
		}
	}
}
