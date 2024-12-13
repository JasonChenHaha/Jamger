package jhttp

import (
	"fmt"
	"jconfig"
	"jlog"
	"time"

	"net/http"
)

type Http struct{}

// ------------------------- outside -------------------------

func NewHttp() *Http {
	htp := &Http{}
	go func() {
		addr := jconfig.GetString("http.addr")
		mux := http.NewServeMux()
		mux.HandleFunc("/", htp.handler)
		server := &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  time.Duration(jconfig.GetInt("http.rTimeout")) * time.Second,
			WriteTimeout: time.Duration(jconfig.GetInt("http.sTimeout")) * time.Second,
		}
		jlog.Info("listen on ", addr)
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return htp
}

// ------------------------- inside -------------------------

func (htp *Http) handler(w http.ResponseWriter, r *http.Request) {
	// 解析url参数
	params := make(map[string]string)
	query := r.URL.Query()
	for k, v := range query {
		params[k] = v[0]
	}
	if r.Method == "GET" {
		fmt.Fprint(w, "this is http server[get]")
	} else {
		fmt.Fprint(w, "this is post server[post]")
	}
}
