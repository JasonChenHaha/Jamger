package jweb

import (
	"jconfig"
	"jglobal"
	"jlog"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type Ses struct {
	id       uint64
	web      *Web
	con      *websocket.Conn
	user     jglobal.User1
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *jglobal.Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(web *Web, con *websocket.Conn, id uint64) *Ses {
	o := &Ses{
		id:       id,
		web:      web,
		con:      con,
		rTimeout: time.Duration(jconfig.GetInt("tcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("tcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *jglobal.Pack, 4),
		qChan:    make(chan any, 2),
	}
	if jconfig.GetBool("noDelay") {
		o.con.UnderlyingConn().(*net.TCPConn).SetNoDelay(true)
	}
	return o
}

func (o *Ses) run() {
	go o.recvGoro()
	go o.sendGoro()
}

func (o *Ses) send(pack *jglobal.Pack) {
	o.sChan <- pack
}

func (o *Ses) close() {
	o.qChan <- 0
}

// ------------------------- inside -------------------------

func (o *Ses) recvGoro() {
	for {
		if o.rTimeout > 0 {
			o.con.SetReadDeadline(time.Now().Add(o.rTimeout))
		}
		_, data, err := o.con.ReadMessage()
		if err != nil {
			jlog.Error(err)
			o.web.Close(o.id)
			return
		}
		pack := &jglobal.Pack{Data: data}
		if err = decoder(pack); err != nil {
			jlog.Error(err)
			o.web.Close(o.id)
			return
		}
		o.user = pack.Ctx.(jglobal.User1)
		o.user.SetSesId(jglobal.WEB, o.id)
		o.web.receive(o.id, pack)
	}
}

func (o *Ses) sendGoro() {
	defer func() {
		err := o.con.Close()
		if o.user != nil {
			o.user.Destory()
		}
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
			if err := encoder(pack); err != nil {
				jlog.Error(err)
				o.web.Close(o.id)
				return
			}
			err := o.con.WriteMessage(websocket.BinaryMessage, pack.Data.([]byte))
			if err != nil {
				jlog.Error(err)
				o.web.Close(o.id)
				return
			}
		}
	}
}
