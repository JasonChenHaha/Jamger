package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"jlog"
	"jpb"
	"net/http"
	"os"
)

type Https struct {
	client *http.Client
}

func testHttps() {
	jlog.Info("<test https>")

	cert, err := os.ReadFile("../template/cert.pem")
	if err != nil {
		jlog.Fatal(err)
	}
	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM(cert)

	tlsConfig := &tls.Config{
		RootCAs: cp,
	}

	htps := &Https{
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		}},
	}
	data := map[string]any{
		"cmd":  jpb.CMD_SIGN_IN_REQ,
		"code": "123",
	}
	rsp := htps.send(data)
	jlog.Debug(rsp)
}

func (htp *Https) send(data map[string]any) map[string]any {
	body, err := json.Marshal(data)
	if err != nil {
		jlog.Fatal(err)
	}
	rsp, err := http.Post("https://"+httpsAddr, "", bytes.NewBuffer(body))
	if err != nil {
		jlog.Fatal(err)
	}
	defer rsp.Body.Close()
	body, err = io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Fatal(err)
	}
	if len(body) == 0 {
		jlog.Fatal("body is empty")
	}
	res := map[string]any{}
	if err = json.Unmarshal(body, &res); err != nil {
		jlog.Fatal(err)
	}
	return res
}
