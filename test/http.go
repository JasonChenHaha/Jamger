package main

import (
	"fmt"
	"io"
	"jconfig"
	"jlog"
	"net/http"
)

func testHttp() {
	jlog.Info("<test http>")
	addr := jconfig.GetString("http.addr")
	rsp, _ := http.Get(fmt.Sprintf("http://%s?abc=1&ddd=2&haha=3", addr))
	defer rsp.Body.Close()

	body, _ := io.ReadAll(rsp.Body)
	jlog.Info(string(body))
}
