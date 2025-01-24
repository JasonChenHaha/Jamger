package jhttp

import (
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Func func(pack *jglobal.Pack)

type Handler struct {
	fun Func
	msg proto.Message
}

type Http struct {
	handler map[jpb.CMD]*Handler
}

// ------------------------- outside -------------------------

func NewHttp() *Http {
	return &Http{}
}

func (htp *Http) AsServer() *Http {
	htp.handler = map[jpb.CMD]*Handler{}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", htp.receive)
		mux.HandleFunc("/auth", htp.authReceive)
		server := &http.Server{
			Addr:    jconfig.GetString("http.addr"),
			Handler: mux,
		}
		jlog.Info("listen on ", jconfig.GetString("http.addr"))
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return htp
}

func (htp *Http) AsClient() *Http {
	return htp
}

func (htp *Http) Register(cmd jpb.CMD, fun Func, msg proto.Message) {
	htp.handler[cmd] = &Handler{
		fun: fun,
		msg: msg,
	}
}

// ------------------------- inside -------------------------

func (htp *Http) authReceive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{W: w}
	if err = decodeRsaToPack(pack, body); err != nil {
		jlog.Warn(err)
		return
	}
	han := htp.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.fun(pack)
	} else {
		han = htp.handler[jpb.CMD_PROXY]
		if han == nil {
			jlog.Error("no proxy cmd.")
			return
		}
		han.fun(pack)
	}
	if o, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(o)
		if err != nil {
			jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	if err = encodePack(pack); err != nil {
		jlog.Error(err)
		return
	}
	if _, err = pack.W.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}

func (htp *Http) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{W: w}
	if err := decodeAesToPack(pack, body); err != nil {
		jlog.Warn(err)
		return
	}
	han := htp.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.msg)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		han.fun(pack)
	} else {
		han = htp.handler[jpb.CMD_PROXY]
		if han == nil {
			jlog.Error("no proxy cmd.")
			return
		}
		han.fun(pack)
	}
	if o, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(o)
		if err != nil {
			jlog.Errorf("%s, cmd: %d", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	if err = encodePack(pack); err != nil {
		jlog.Error(err)
		return
	}
	if _, err = pack.W.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}
