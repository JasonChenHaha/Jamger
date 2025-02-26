package jweb

import (
	"jconfig"
	"jlog"
	"jpb"
	"time"

	"github.com/gorilla/websocket"
)

type Ses struct {
	web      *Web
	con      *websocket.Conn
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(web *Web, con *websocket.Conn, id uint64) *Ses {
	o := &Ses{
		web:      web,
		con:      con,
		id:       id,
		rTimeout: time.Duration(jconfig.GetInt("tcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("tcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *Pack, 4),
		qChan:    make(chan any, 2),
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
			_, data, err := o.con.ReadMessage()
			if err != nil {
				o.web.delete(o.id)
				return
			}
			pack := unserializeData(data)
			switch pack.Cmd {
			case jpb.CMD_HEARTBEAT:
			default:
				o.web.receive(o.id, pack)
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
			data := serializePack(pack)
			err := o.con.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				jlog.Error(err)
				return
			}
		}
	}
}
