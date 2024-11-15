package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"jamger/comm"
	jlog "jamger/log"
	"net"
	"sync"
	"time"
)

const (
	HEAD_SIZE = 2
	CMD_SIZE  = 2
)

type Pack struct {
	Cmd  uint16
	Data []byte
}

type Session struct {
	id       int64
	con      net.Conn
	sendQ    comm.Queue
	rTimeout time.Duration
	wTimeout time.Duration
	callback func(sesID int64, pack Pack)
	wait     sync.WaitGroup
}

func (ses *Session) run() {
	ses.wait.Add(2)
	go ses.recvGoro()
	go ses.sendGoro()
	go func() {
		ses.wait.Wait()
		ses.con.Close()
		g_sesMgr.delete(ses.id)
	}()
}

func (ses *Session) recvGoro() {
	for {
		if ses.rTimeout > 0 {
			ses.con.SetReadDeadline(time.Now().Add(ses.rTimeout))
		}
		pack, err := ses.recvPack()
		if err != nil {
			jlog.Errorln("session recv error: ", err)
			break
		}
		ses.callback(ses.id, pack)
	}
	ses.wait.Done()
}

func (ses *Session) sendGoro() {
	for {
		if ses.wTimeout > 0 {
			ses.con.SetWriteDeadline(time.Now().Add(ses.wTimeout))
		}
		packs := ses.sendQ.PickAll().([]Pack)
		for _, p := range packs {
			err := ses.sendPack(p)
			if err != nil {
				jlog.Errorln("session send error: ", err)
				break
			}
		}
	}
	ses.wait.Done()
}

func (ses *Session) recvPack() (pack Pack, err error) {
	buffer := make([]byte, HEAD_SIZE)
	_, err = io.ReadFull(ses.con, buffer)
	if err != nil {
		return
	}
	size := binary.LittleEndian.Uint16(buffer)
	if size < HEAD_SIZE+CMD_SIZE {
		err = fmt.Errorf("pack size invalid")
		return
	}
	buffer = make([]byte, size)
	_, err = io.ReadFull(ses.con, buffer)
	if err != nil {
		return
	}
	pack.Cmd = binary.LittleEndian.Uint16(buffer)
	pack.Data = buffer[CMD_SIZE:]
	return
}

func (ses *Session) sendPack(pack Pack) error {
	size := HEAD_SIZE + CMD_SIZE + len(pack.Data)
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(size))
	binary.LittleEndian.PutUint16(buffer[HEAD_SIZE:], pack.Cmd)
	copy(buffer[HEAD_SIZE+CMD_SIZE:], pack.Data)
	for pos := 0; pos < size; {
		n, err := ses.con.Write(buffer)
		if err != nil {
			break
		}
		pos += n
	}
	return nil
}

// --------------------------- interface ---------------------------

func (ses *Session) send(pack Pack) {
	ses.sendQ = append(ses.sendQ, pack)
}
