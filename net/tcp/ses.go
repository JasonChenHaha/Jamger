package jtcp

import (
	"encoding/binary"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"net"
	"time"
)

const (
	packSize = 2
)

type Ses struct {
	id       uint64
	tcp      *Tcp
	con      *net.TCPConn
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *jglobal.Pack
	qChan    chan any
}

// ------------------------- package -------------------------

func newSes(tcp *Tcp, con net.Conn, id uint64) *Ses {
	ses := &Ses{
		id:       id,
		tcp:      tcp,
		con:      con.(*net.TCPConn),
		rTimeout: time.Duration(jconfig.GetInt("tcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("tcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *jglobal.Pack, 4),
		qChan:    make(chan any, 2),
	}
	if jconfig.GetBool("noDelay") {
		ses.con.SetNoDelay(true)
	}
	return ses
}

func (ses *Ses) run() {
	go ses.recvGoro()
	go ses.sendGoro()
}

func (ses *Ses) send(pack *jglobal.Pack) {
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
			data, err := ses.recvBytes()
			if err != nil {
				ses.tcp.delete(ses.id)
				return
			}
			pack := &jglobal.Pack{Data: data}
			err = decoder(pack)
			if err != nil {
				ses.tcp.delete(ses.id)
				return
			}
			ses.tcp.receive(ses, pack)
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
			if err := encoder(pack); err != nil {
				jlog.Error(err)
				ses.tcp.delete(ses.id)
				return
			}
			if err := ses.sendBytes(pack); err != nil {
				jlog.Error(err)
				ses.tcp.delete(ses.id)
			}
		}
	}
}

func (ses *Ses) recvBytes() ([]byte, error) {
	raw := make([]byte, packSize)
	if _, err := io.ReadFull(ses.con, raw); err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint16(raw)
	raw = make([]byte, size)
	if _, err := io.ReadFull(ses.con, raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (ses *Ses) sendBytes(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	size := len(data)
	raw := make([]byte, packSize+size)
	binary.LittleEndian.PutUint16(raw, uint16(size))
	copy(raw[packSize:], raw)
	for pos := 0; pos < size; {
		n, err := ses.con.Write(raw)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
