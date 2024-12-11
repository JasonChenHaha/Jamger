package main

import (
	"io"
	"jlog"
	"net/http"
)

func testHttp() {
	jlog.Info("<test http>")
	rsp, _ := http.Get("http://127.0.0.1:8080?abc=1&ddd=2&haha=3")
	defer rsp.Body.Close()

	body, _ := io.ReadAll(rsp.Body)
	jlog.Info(string(body))
}
