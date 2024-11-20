package main

import (
	"io"
	jlog "jamger/log"
	"net/http"
)

func testHttp() {
	jlog.Info("<test http>")
	rsp, err := http.Get("http://127.0.0.1:8080?abc=1&ddd=2&haha=3")
	if err != nil {
		jlog.Fatal(err)
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Fatal(err)
	}
	jlog.Info(string(body))
}
