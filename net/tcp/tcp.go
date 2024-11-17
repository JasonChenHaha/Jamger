package tcp

import (
	jconfig "jamger/config"
	jlog "jamger/log"
	"net"
)

// ------------------------- outside -------------------------

func Run() {
	cfg := jconfig.Get("tcp").(map[string]any)
	addr := cfg["addr"].(string)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		jlog.Panic(err)
	}
	jlog.Debug("listen ok!!!")

	go accept(listener)
}

// ------------------------- inside -------------------------

func accept(listener net.Listener) {
	for {
		con, err := listener.Accept()
		if err != nil {
			jlog.Error(err)
			continue
		} else {
			g_sesMgr.add(con)
		}
	}
}
