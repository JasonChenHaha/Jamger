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

type Ses struct {
	id       uint64
	tcp      *Tcp
	con      *net.TCPConn
	user     jglobal.User1
	rTimeout time.Duration
	sTimeout time.Duration
	sChan    chan *jglobal.Pack
	qChan    chan any
}

const (
	packSize = 4
)

// ------------------------- package -------------------------

func newSes(tcp *Tcp, con net.Conn, id uint64) *Ses {
	o := &Ses{
		id:       id,
		tcp:      tcp,
		con:      con.(*net.TCPConn),
		rTimeout: time.Duration(jconfig.GetInt("tcp.rTimeout")) * time.Millisecond,
		sTimeout: time.Duration(jconfig.GetInt("tcp.sTimeout")) * time.Millisecond,
		sChan:    make(chan *jglobal.Pack, 4),
		qChan:    make(chan any, 2),
	}
	if jconfig.GetBool("noDelay") {
		o.con.SetNoDelay(true)
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
		data, err := o.recvBytes()
		if err != nil {
			jlog.Error(err)
			o.tcp.Close(o.id)
			return
		}
		pack := &jglobal.Pack{Data: data}
		if err = decoder(pack); err != nil {
			jlog.Error(err)
			o.tcp.Close(o.id)
			return
		}
		o.user = pack.Ctx.(jglobal.User1)
		o.user.SetSesId(jglobal.TCP, o.id)
		o.tcp.receive(o.id, pack)
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
				o.tcp.Close(o.id)
				return
			}
			if err := o.sendBytes(pack); err != nil {
				jlog.Error(err)
				o.tcp.Close(o.id)
			}
		}
	}
}

func (o *Ses) recvBytes() ([]byte, error) {
	raw := make([]byte, packSize)
	if _, err := io.ReadFull(o.con, raw); err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint32(raw)
	raw = make([]byte, size)
	if _, err := io.ReadFull(o.con, raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (o *Ses) sendBytes(pack *jglobal.Pack) error {
	data := pack.Data.([]byte)
	size := len(data)
	raw := make([]byte, packSize+size)
	binary.LittleEndian.PutUint32(raw, uint32(size))
	copy(raw[packSize:], data)
	for pos := 0; pos < size; {
		n, err := o.con.Write(raw)
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
