package tcp

import (
	jlog "jamger/log"
	"net"
	"sync"
	"sync/atomic"
)

type SessionMgr struct {
	sesIdCounter uint64
	ses          sync.Map
}

var g_sesMgr = SessionMgr{}

// ------------------------- package -------------------------

func (mgr *SessionMgr) add(con net.Conn) {
	id := atomic.AddUint64(&mgr.sesIdCounter, 1)
	ses := newSession(id)
	mgr.ses.Store(id, ses)
	ses.run(con)
}

func (mgr *SessionMgr) close(id uint64) {
	obj, ok := mgr.ses.Load(id)
	if ok {
		mgr.ses.Delete(id)
		obj.(*Session).close()
	}
}

func (mgr *SessionMgr) send(id uint64, pack Pack) {
	obj, ok := mgr.ses.Load(id)
	if !ok {
		jlog.Errorf("session %d not found", id)
		return
	}
	obj.(*Session).send(pack)
}
