package juBase

// 保护模块用来解决缓存在多节点同时存在时的数据一致性问题：
// 若group集群内有a，b两个节点,当a节点下线后，请求从gate转发到b节点，b节点会建立缓存，
// 此时a，b节点都缓存了同一份数据，可能会导致将旧数据写入DB。
// 解决方法是当节点加入group集群时，一段时间内进入保护模式（时间和user的淘汰时间一致）
// 在保护模式内，首次收到user的请求时，找到节点加入前该user路由到的旧节点，给它发通知
// 让其删除user缓存。

import (
	"jglobal"
	"jpb"
	"jrpc"
	"jschedule"
)

type Protect struct {
	rec map[uint32]struct{}
}

var prot *Protect

// ------------------------- outside -------------------------

func CreateProtect() {
	prot = &Protect{}
	jschedule.DoAt(EXPIRE, func() {
		prot = nil
	})
}

func IsProtectMode(uid uint32) bool {
	if prot != nil {
		if _, ok := prot.rec[uid]; !ok {
			// 如果两次映射的节点不同，说明发生了路由变动
			if target := jrpc.GetLastConsistentHashTarget(uid); target != jrpc.GetConsistentHashTarget(jglobal.GROUP, uid) {
				go target.Call(&jglobal.Pack{
					Cmd:  jpb.CMD_DEL_USER,
					Data: &jpb.DelUserReq{Uid: uid},
				}, &jpb.DelUserRsp{})
			}
			prot.rec[uid] = struct{}{}
			return true
		}
	}
	return false
}
