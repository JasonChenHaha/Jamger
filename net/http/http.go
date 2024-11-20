package jhttp

import (
	"fmt"
	jconfig "jamger/config"
	jlog "jamger/log"
	"time"

	"net/http"
)

type Http struct{}

// ------------------------- outside -------------------------

func NewHttp() *Http {
	return &Http{}
}

func (htp *Http) Run() {
	go func() {
		addr := jconfig.Get("http.addr").(string)
		mux := http.NewServeMux()
		mux.HandleFunc("/", htp.handler)
		server := &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  time.Duration(jconfig.Get("http.rTimeout").(int)) * time.Second,
			WriteTimeout: time.Duration(jconfig.Get("http.sTimeout").(int)) * time.Second,
		}
		jlog.Info("listen on ", addr)
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
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
