package jhttp

import (
	"fmt"
	jconfig "jamger/config"
	jlog "jamger/log"

	"net/http"
)

type Http struct{}

// ------------------------- outside -------------------------

func NewHttp() *Http {
	return &Http{}
}

func (htp *Http) Run() {
	go func() {
		cfg := jconfig.Get("http").(map[string]any)
		addr := cfg["addr"].(string)
		http.HandleFunc("/", htp.handler)
		jlog.Info("listen on ", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
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
		fmt.Fprint(w, "this is get server")
	} else {
		fmt.Fprint(w, "this is post server")
	}
}
