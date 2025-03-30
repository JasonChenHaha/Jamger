package jetcd

import (
	"context"
	"fmt"
	"jconfig"
	"jglobal"
	"jlog"
	"jschedule"
	"juser"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Handler func(group int, index int, info map[string]any) // handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件

type etcd struct {
	*clientv3.Client
	lease   *clientv3.LeaseGrantResponse
	info    map[string]any                 // 本服信息
	server  map[int]map[int]map[string]any // map[group][index] = infos
	watch   map[int][2]Handler             // map[group] = [joinHandler, leaveHandler]
	ticker  any
	ticker2 any
}

var Etcd *etcd

// ------------------------- outside -------------------------

func Init() {
	Etcd = &etcd{
		info:   map[string]any{"addr": jconfig.GetString("rpc.addr")},
		server: map[int]map[int]map[string]any{},
		watch:  map[int][2]Handler{},
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{jconfig.GetString("etcd.addr")}})
	if err != nil {
		jlog.Fatal(err)
	}
	Etcd.Client = cli
	// 延时执行是为了等待watch执行
	jschedule.DoAt(time.Second, Etcd.beginProcress)
}

func (o *etcd) Watch(group int, join Handler, leave Handler) {
	o.watch[group] = [2]Handler{join, leave}
}

// ------------------------- inside -------------------------

// 起服流程
// (解决一致性哈希路由+集群缓存情况下的，节点下线、网络分区、动态扩缩容时，数据一致性问题)
// 首先从etcd拉取本集群列表，如果集群列表
// 为空
//		说明集群整体启动
// 不为空
//		说明本节点加入集群，如果集群列表中本节点
//		存在
//			说明本节点不久前刚上报过，集群路由没有发生变化，这时再检查从etcd拉下来的本节点
// 			信息中是否存在protect过期时间，存在并且时间还未到，就切换为protect模式，时间
// 			到了就将其从etcd中删除。
//		不存在
//			说明集群路由已经变化，原本路由过来的请求现在已经路由到其他节点并生成缓存
//			这时需要切换为protect模式，将恢复后路由过来的请求，通过比对新旧两份集群列表
// 			来得出该请求前后路由到的节点差异，以此为依据来通知指定节点卸载缓存，protect
// 			模式持续时间应该和缓存过期淘汰时间一致，并且将时间写入etcd来保存状态（否则本
// 			节点此时重启发现路由表存在本节点，就无法正确判断是否需要切换protect模式）。

func (o *etcd) beginProcress(args ...any) {
	if Etcd.ticker != nil {
		jschedule.Stop(Etcd.ticker)
		Etcd.ticker = nil
	}
	if Etcd.ticker2 != nil {
		jschedule.Stop(Etcd.ticker2)
		Etcd.ticker2 = nil
	}
	Etcd.lease = nil
again:
	if o.download(true) == context.DeadlineExceeded {
		goto again
	}
	// protect模式判断
	if v := o.server[jglobal.GROUP]; v != nil {
		if v2 := v[jglobal.INDEX]; v2 == nil {
			// 本节点是新加入进来的，切换protect模式
			juser.Protect.Enable(true)
			o.info["protect"] = time.Now().Unix() + jglobal.USER_LIVE
			o.lease = nil
			o.ticker2 = jschedule.DoAt(jglobal.USER_LIVE*time.Second, func(args ...any) {
				juser.Protect.Disable()
				o.info["protect"] = nil
				o.lease = nil
			})
		} else {
			if v3, ok := v2["protect"]; ok {
				// protect模式还未结束
				left := int64(v3.(float64)) - time.Now().Unix()
				if left > 0 {
					juser.Protect.Enable(false)
					o.ticker2 = jschedule.DoAt(time.Duration(left)*time.Second, func(args ...any) {
						juser.Protect.Disable()
						delete(o.info, "protect")
						o.lease = nil
					})
				} else {
					juser.Protect.Disable()
					delete(o.info, "protect")
					o.lease = nil
				}
			}
		}
	}
	// 为了防止上一步出现误判
	time.Sleep(3 * time.Second)
again2:
	if o.upload() == context.DeadlineExceeded {
		goto again2
	}
	// 为了下一步的download能获取完整集群信息
	time.Sleep(3 * time.Second)
again3:
	// 再次download是为了集群启动时快速建立内部网络
	if o.download(false) == context.DeadlineExceeded {
		goto again3
	}
	Etcd.ticker = jschedule.DoEvery(time.Duration(jconfig.GetInt("etcd.update"))*time.Millisecond, Etcd.tick)
}

// 上报
func (o *etcd) upload() error {
	if o.lease == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Millisecond)
		defer cancel()
		lease, err := o.Grant(ctx, int64(jconfig.GetInt("etcd.keepalive")/1000))
		if err != nil {
			jlog.Error(err)
			return err
		}
		o.lease = lease
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Millisecond)
		defer cancel2()
		key := fmt.Sprintf("/%d/%d/%d", jglobal.ZONE, jglobal.GROUP, jglobal.INDEX)
		_, err = o.Put(ctx2, key, string(jglobal.SerializeJson(o.info)), clientv3.WithLease(o.lease.ID))
		if err != nil {
			jlog.Error(err)
			return err
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Millisecond)
		defer cancel()
		_, err := o.KeepAliveOnce(ctx, o.lease.ID)
		if err != nil {
			jlog.Error(err)
			return err
		}
	}
	return nil
}

// 下载
func (o *etcd) download(init bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(jconfig.GetInt("etcd.timeout"))*time.Millisecond)
	defer cancel()
	rsp, err := o.Get(ctx, fmt.Sprintf("/%d/", jglobal.ZONE), clientv3.WithPrefix())
	if err != nil {
		jlog.Error(err)
		return err
	}
	tmp := map[int]map[int]map[string]any{}
	for _, kv := range rsp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		group, index := jglobal.Atoi[int](parts[2]), jglobal.Atoi[int](parts[3])
		if tmp[group] == nil {
			tmp[group] = map[int]map[string]any{}
		}
		var info map[string]any
		jglobal.UnserializeJson(kv.Value, &info)
		if group == jglobal.GROUP && index == jglobal.INDEX && init {
			o.info = info
		}
		tmp[group][index] = info
		if o.server[group] == nil || o.server[group][index] == nil {
			// join
			if _, ok := o.watch[group]; ok {
				o.watch[group][0](group, index, info)
			}
		}
	}
	for group, v := range o.server {
		for index, info := range v {
			if tmp[group] == nil || tmp[group][index] == nil {
				// leave
				if _, ok := o.watch[group]; ok {
					o.watch[group][1](group, index, info)
				}
			}
		}
	}
	o.server = tmp
	return nil
}

func (o *etcd) tick(args ...any) {
	// 当发生超时时，本节点无法确定etcd的租约是否已经过期，因此需要按最坏的可能来应对
	// 走起服流程来判断是否需要进入protect模式
	if o.upload() == context.DeadlineExceeded {
		// 一旦发生超时，立刻进入启服流程
		o.beginProcress()
		return
	}
	if o.download(false) == context.DeadlineExceeded {
		// 一旦发生超时，立刻进入启服流程
		o.beginProcress()
		return
	}
}
