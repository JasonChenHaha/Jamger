package jhttps

import (
	"encoding/json"
	"fmt"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

type Https struct {
	mux     *http.ServeMux
	handler map[jpb.CMD]*Handler
}

type Handler struct {
	fun      func(*jglobal.Pack)
	template proto.Message
}

var encoder func(*jglobal.Pack) error
var decoder func(string, *jglobal.Pack) error

// ------------------------- outside -------------------------

func NewHttps() *Https {
	return &Https{}
}

func (o *Https) AsServer() *Https {
	o.handler = map[jpb.CMD]*Handler{}
	go func() {
		o.mux = http.NewServeMux()
		o.mux.HandleFunc("/", o.receive)
		o.mux.HandleFunc("/image/", o.imageReceive)
		o.mux.HandleFunc("/video/", o.videoReceive)
		o.mux.HandleFunc("/test/", o.test)
		server := &http.Server{
			Addr:    jconfig.GetString("https.addr"),
			Handler: o.mux,
		}
		jlog.Info("listen on ", jconfig.GetString("https.addr"))
		// if err := server.ListenAndServeTLS(jconfig.GetString("https.crt"), jconfig.GetString("https.key")); err != nil {
		// 	jlog.Fatal(err)
		// }
		if err := server.ListenAndServe(); err != nil {
			jlog.Fatal(err)
		}
	}()
	return o
}

func (o *Https) AsClient() *Https {
	return o
}

func (o *Https) SetCodec(en func(*jglobal.Pack) error, de func(string, *jglobal.Pack) error) {
	encoder = en
	decoder = de
}

func (o *Https) Register(cmd jpb.CMD, fun func(*jglobal.Pack), template proto.Message) {
	o.handler[cmd] = &Handler{
		fun:      fun,
		template: template,
	}
}

func (o *Https) Get(url string) (map[string]any, error) {
	rsp, err := http.Get(url)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	res := map[string]any{}
	if err = json.Unmarshal(body, &res); err != nil {
		jlog.Error(err)
		return nil, err
	}
	return res, nil
}

// ------------------------- inside -------------------------

func (o *Https) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{Data: body}
	if err = decoder(r.URL.Path, pack); err != nil {
		return
	}
	han := o.handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.template)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.fun(pack)
	} else {
		han = o.handler[jpb.CMD_TRANSFER]
		if han == nil {
			jlog.Error("no cmd(TRANSFER).")
			return
		}
		han.fun(pack)
	}
	if v, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(v)
		if err != nil {
			jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	if err = encoder(pack); err != nil {
		return
	}
	if _, err = w.Write(pack.Data.([]byte)); err != nil {
		jlog.Error(err)
	}
}

func (o *Https) imageReceive(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	pack := &jglobal.Pack{
		Cmd:  jpb.CMD_IMAGE_REQ,
		Data: &jpb.ImageReq{Uid: jglobal.Atoi[uint32](parts[len(parts)-1])},
	}
	han := o.handler[pack.Cmd]
	if han == nil {
		jlog.Errorf("no cmd(%s)", pack.Cmd)
		return
	}
	han.fun(pack)
	if v, ok := pack.Data.(*jpb.ImageRsp); ok {
		if _, err := w.Write(v.Image); err != nil {
			jlog.Error(err)
		}
	}
}

func (o *Https) videoReceive(w http.ResponseWriter, r *http.Request) {
	jlog.Debug(r.Header)
	parts := strings.Split(r.URL.Path, "/")
	if r.Header["Range"] == nil {
		pack := &jglobal.Pack{
			Cmd: jpb.CMD_VIDEO_REQ,
			Data: &jpb.VideoReq{
				Uid:   jglobal.Atoi[uint32](parts[len(parts)-1]),
				Start: 0,
				End:   math.MaxUint32 - 1,
			},
		}
		han := o.handler[pack.Cmd]
		if han == nil {
			jlog.Errorf("no cmd(%s)", pack.Cmd)
			return
		}
		han.fun(pack)
		rsp := pack.Data.(*jpb.VideoRsp)
		size := len(rsp.Video)
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", size-1, rsp.Size))
		w.WriteHeader(http.StatusPartialContent)
		if _, err := w.Write(rsp.Video); err != nil {
			jlog.Error(err)
		}
	} else {
		ab := strings.Split(strings.Split(r.Header["Range"][0], "=")[1], "-")
		if ab[1] == "" {
			if ab[0] == "0" {
				pack := &jglobal.Pack{
					Cmd: jpb.CMD_VIDEO_REQ,
					Data: &jpb.VideoReq{
						Uid:   jglobal.Atoi[uint32](parts[len(parts)-1]),
						Start: 0,
						End:   1048575,
					},
				}
				han := o.handler[pack.Cmd]
				if han == nil {
					jlog.Errorf("no cmd(%s)", pack.Cmd)
					return
				}
				han.fun(pack)
				// if pack.code panduan
				rsp := pack.Data.(*jpb.VideoRsp)
				size := len(rsp.Video)
				// w.Header().Set("Content-Type", "video/mp4")
				// w.Header().Set("Accept-Ranges", "bytes")
				// w.Header().Set("Connection", "keep-alive")
				// w.Header().Set("Content-Length", jglobal.Itoa(size))
				w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", size-1, rsp.Size))
				w.WriteHeader(http.StatusPartialContent)
				if _, err := w.Write(rsp.Video); err != nil {
					jlog.Error(err)
				}
			} else {
				start := jglobal.Atoi[uint32](ab[0])
				pack := &jglobal.Pack{
					Cmd: jpb.CMD_VIDEO_REQ,
					Data: &jpb.VideoReq{
						Uid:   jglobal.Atoi[uint32](parts[len(parts)-1]),
						Start: start,
						End:   1048575,
					},
				}
				han := o.handler[pack.Cmd]
				if han == nil {
					jlog.Errorf("no cmd(%s)", pack.Cmd)
					return
				}
				han.fun(pack)
				rsp := pack.Data.(*jpb.VideoRsp)
				size := uint32(len(rsp.Video))
				// w.Header().Set("Content-Length", jglobal.Itoa(size))
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start+size-1, rsp.Size))
				w.WriteHeader(http.StatusPartialContent)
				if _, err := w.Write(rsp.Video); err != nil {
					jlog.Error(err)
				}
			}
		} else {
			start := jglobal.Atoi[uint32](ab[0])
			end := jglobal.Atoi[uint32](ab[1])
			pack := &jglobal.Pack{
				Cmd: jpb.CMD_VIDEO_REQ,
				Data: &jpb.VideoReq{
					Uid:   jglobal.Atoi[uint32](parts[len(parts)-1]),
					Start: start,
					End:   end,
				},
			}
			han := o.handler[pack.Cmd]
			if han == nil {
				jlog.Errorf("no cmd(%s)", pack.Cmd)
				return
			}
			han.fun(pack)
			rsp := pack.Data.(*jpb.VideoRsp)
			size := uint32(len(rsp.Video))
			w.Header().Set("Content-Length", jglobal.Itoa(size))
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start+size-1, rsp.Size))
			w.WriteHeader(http.StatusPartialContent)
			if _, err := w.Write(rsp.Video); err != nil {
				jlog.Error(err)
			}
		}
	}

	// parts := strings.Split(r.URL.Path, "/")
	// pack := &jglobal.Pack{
	// 	Cmd:  jpb.CMD_VIDEO_REQ,
	// 	Data: &jpb.VideoReq{Uid: jglobal.Atoi[uint32](parts[len(parts)-1])},
	// }
	// han := o.handler[pack.Cmd]
	// if han == nil {
	// 	jlog.Errorf("no cmd(%s)", pack.Cmd)
	// 	return
	// }
	// han.fun(pack)

	// w.Header().Set("Content-Type", "video/mp4")
	// w.Header().Set("Accept-Ranges", "bytes")
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	// http.ServeContent(w, r, "video.mp4", time.Now(), reader)
}

func (o *Https) test(w http.ResponseWriter, r *http.Request) {
	jlog.Debug(r.Header)

	file, err := os.Open("../../template/abc.mp4")
	if err != nil {
		jlog.Error(err)
		return
	}

	info, err := file.Stat()
	if err != nil {
		jlog.Error(err)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}
