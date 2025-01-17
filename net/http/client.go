package jhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"jconfig"
	"jlog"
	"jpb"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

type HttpClient struct {
	addr   string
	client *http.Client
}

// ------------------------- outside -------------------------

func NewHttpClient(addr string) *HttpClient {
	hc := &HttpClient{
		addr: "http://" + addr,
		client: &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, network, addr)
				},
			},
			Timeout: time.Duration(jconfig.GetInt("http.timeout")) * time.Millisecond,
		},
	}
	return hc
}

func (hc *HttpClient) Send(cmd jpb.CMD, msg []byte) *Pack {
	raw := SerializePack(&Pack{
		Cmd:  cmd,
		Data: msg,
	})
	rsp, err := hc.client.Post(hc.addr, "", bytes.NewBuffer(raw))
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return nil
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Errorf("%s, %d", err, cmd)
		return nil
	}
	return UnserializeToPack(body)
}
