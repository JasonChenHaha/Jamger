package juBase

// 保护模块用来解决缓存在多节点同时存在时的数据一致性问题：
// 若group集群内有a，b两个节点,当a节点下线后，请求从gate转发到b节点，b节点会建立缓存，
// 此时a，b节点都缓存了同一份数据，可能会导致将旧数据写入DB。
// 解决方法是当节点加入group集群时，一段时间内进入保护模式（时间和user的淘汰时间一致）
// 在保护模式内，首次收到user的请求时，找到节点加入前该user路由到的旧节点，给它发通知
// 让其删除user缓存。

import (
	"jglobal"
	"jnrpc"
	"jpb"
	"sync"
)

type Rpc interface {
	GetLastConsistentHashTarget(uint32) *jnrpc.Rpc
	GetConsistentHashTarget(int, uint32) *jnrpc.Rpc
}

type protect struct {
	rpc    Rpc
	enable bool
	rec    map[uint32]struct{}
	ticker any
	mutex  sync.RWMutex
}

var Protect *protect

// ------------------------- outside -------------------------

func Init(rpc Rpc) {
	Protect = &protect{
		rpc: rpc,
		rec: map[uint32]struct{}{},
	}
}

func (o *protect) Enable(new bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.enable = true
	// 四种情况导致enable
	// 1.起服时, new enable
	// 2.断网恢复时, new enable
	// 3.断网恢复-再断网恢复（租约过期）new enable
	// 4.断网恢复-再断网恢复（租约未过期）continue enable，此时要保留rec记录，防止重复notify
	if new {
		o.rec = map[uint32]struct{}{}
	} else {
		o.rec = map[uint32]struct{}{}
	}
}

func (o *protect) Disable() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.enable = false
	o.rec = map[uint32]struct{}{}
}

func (o *protect) Touch(uid uint32) bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	if o.enable {
		if _, ok := o.rec[uid]; !ok {
			// 如果两次映射的节点不同，说明发生了路由变动
			if target := o.rpc.GetLastConsistentHashTarget(uid); target != o.rpc.GetConsistentHashTarget(jglobal.GROUP, uid) {
				go target.Call(&jglobal.Pack{
					Cmd:  jpb.CMD_DEL_USER,
					Data: &jpb.DeleteUserReq{Uid: uid},
				}, &jpb.DeleteUserRsp{})
			}
			o.rec[uid] = struct{}{}
			return true
		}
	}
	return false
}
