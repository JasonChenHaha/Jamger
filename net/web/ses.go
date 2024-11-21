package jweb

import (
	jconfig "jamger/config"
	jlog "jamger/log"

	"github.com/gorilla/websocket"
)

type Ses struct {
	web   *Web
	con   *websocket.Conn
	id    uint64
	sChan chan *Pack
	qChan chan any
}

// ------------------------- package -------------------------

func newSes(web *Web, con *websocket.Conn, id uint64) *Ses {
	ses := &Ses{
		web:   web,
		con:   con,
		id:    id,
		sChan: make(chan *Pack, jconfig.GetInt("web.sBufferSize")),
		qChan: make(chan any, 2),
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
			_, data, err := ses.con.ReadMessage()
			if err != nil {
				err := err.(*websocket.CloseError)
				if err.Code != websocket.CloseNormalClosure && err.Code != websocket.CloseAbnormalClosure {
					jlog.Error(err)
				}
				ses.web.delete(ses.id)
				return
			}
			pack := unserializeData(data)
			ses.web.receive(ses.id, pack)
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
			data := serializePack(pack)
			err := ses.con.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				jlog.Error(err)
				return
			}
		}
	}
}
