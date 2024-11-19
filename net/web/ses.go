package jweb

import (
	"io"
	jconfig "jamger/config"
	"jamger/global"
	jlog "jamger/log"
	"time"

	"github.com/gorilla/websocket"
)

type Ses struct {
	web      *Web
	id       uint64
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(web *Web, id uint64) *Ses {
	return &Ses{
		web:      web,
		id:       id,
		rTimeout: time.Duration(jconfig.Get("web.rTimeout").(int)) * time.Second,
		sTimeout: time.Duration(jconfig.Get("web.sTimeout").(int)) * time.Second,
		sChan:    make(chan *Pack, global.G_TCP_SEND_BUFFER_SIZE),
		qChan:    make(chan any, 4),
	}
}

func (ses *Ses) run(con *websocket.Conn) {
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

func (ses *Ses) recvGoro(con *websocket.Conn) {
	defer ses.web.delete(ses.id)
	for {
		select {
		case <-ses.qChan:
			return
		default:
			if ses.rTimeout > 0 {
				con.SetReadDeadline(time.Now().Add(ses.rTimeout))
			}
			_, data, err := con.ReadMessage()
			if err != nil {
				if err != io.EOF {
					jlog.Error(err)
				}
				return
			}
			pack, err := unserializeData(data)
			if err != nil {
				jlog.Error(err)
				return
			}
			ses.web.receive(ses.id, pack)
		}
	}
}

func (ses *Ses) sendGoro(con *websocket.Conn) {
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
			data, err := serializePack(pack)
			if err != nil {
				jlog.Error(err)
				return
			}
			err = con.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				jlog.Error(err)
				return
			}

		}
	}
}
