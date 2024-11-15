package tcp

import (
	"net"
	"sync"
	"sync/atomic"
)

type SessionMgr struct {
	ses   map[int64]*Session
	sesId int64
	lock  sync.RWMutex
}

var g_sesMgr = SessionMgr{}

func (mgr *SessionMgr) add(con net.Conn) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	id := atomic.AddInt64(&mgr.sesId, 1)
	ses := &Session{id: id, con: con}
	g_sesMgr.ses[id] = ses
	ses.run()
}

func (mgr *SessionMgr) delete(id int64) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	delete(mgr.ses, id)
}
