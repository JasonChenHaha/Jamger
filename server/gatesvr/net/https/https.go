package jhttps2

import (
	"encoding/binary"
	"fmt"
	"io"
	"jglobal"
	"jlog"
	"jnet"
	"jpb"
	"net/http"
	"strings"

	"google.golang.org/protobuf/proto"
)

// ------------------------- outside -------------------------

func Init() {
	jnet.Https.RegisterPattern("/auth", authReceive)
	jnet.Https.RegisterPattern("/image/", imageReceive)
	jnet.Https.RegisterPattern("/video/", videoReceive)
}

// ------------------------- inside -------------------------

// auth client https pack structure:
// +--------------------+
// |        pack        |
// +---------+----------+
// |   cmd   |   data   |
// |---------+----------+
// |    2    |   ...    |
// +---------+----------+
// auth server https pack structure:
// +--------------------+
// |        pack        |
// +---------+----------+
// |   cmd   |   data   |
// |---------+----------+
// |    2    |   ...    |
// +---------+----------+
func authReceive(w http.ResponseWriter, r *http.Request) {
	// 解包
	body, err := io.ReadAll(r.Body)
	if err != nil {
		jlog.Error(err)
		return
	}
	pack := &jglobal.Pack{
		Cmd:  jpb.CMD(binary.LittleEndian.Uint16(body)),
		Data: body[jglobal.CMD_SIZE:],
	}
	// 执行
	han := jnet.Https.Handler[pack.Cmd]
	if han != nil {
		msg := proto.Clone(han.Template)
		if err = proto.Unmarshal(pack.Data.([]byte), msg); err != nil {
			jlog.Warnf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = msg
		han.Fun(pack)
	} else {
		han = jnet.Https.Handler[jpb.CMD_TRANSFER]
		if han == nil {
			jlog.Error("no cmd(TRANSFER).")
			return
		}
		han.Fun(pack)
	}
	// 打包
	if v, ok := pack.Data.(proto.Message); ok {
		tmp, err := proto.Marshal(v)
		if err != nil {
			jlog.Errorf("%s, cmd(%s)", err, pack.Cmd)
			return
		}
		pack.Data = tmp
	}
	data := pack.Data.([]byte)
	raw := make([]byte, jglobal.CMD_SIZE+len(data))
	binary.LittleEndian.PutUint16(raw, uint16(pack.Cmd))
	copy(raw[jglobal.CMD_SIZE:], data)
	// 发送
	if _, err = w.Write(raw); err != nil {
		jlog.Error(err)
	}
}

func imageReceive(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	pack := &jglobal.Pack{
		Cmd:  jpb.CMD_IMAGE_REQ,
		Data: &jpb.ImageReq{Uid: jglobal.Atoi[uint32](parts[len(parts)-1])},
	}
	han := jnet.Https.Handler[pack.Cmd]
	if han == nil {
		jlog.Errorf("no cmd(%s)", pack.Cmd)
		return
	}
	han.Fun(pack)
	if v, ok := pack.Data.(*jpb.ImageRsp); ok {
		if _, err := w.Write(v.Image); err != nil {
			jlog.Error(err)
		}
	}
}

func videoReceive(w http.ResponseWriter, r *http.Request) {
	jlog.Debug(r.Header)
	parts := strings.Split(r.URL.Path, "/")
	start, end := uint32(0), uint32(0)
	if r.Header["Range"] == nil {
		// 不分片的初始请求，返回视频元数据和少量长度的视频头部
	} else {
		// 分片请求
		ab := strings.Split(strings.Split(r.Header["Range"][0], "=")[1], "-")
		if ab[1] == "" {
			// 没有给定请求范围右端的情况
			if ab[0] == "0" {
				// 分片的初始请求，返回视频元数据和少量长度的视频头部
			} else {
				// 分片的后续请求，返回适当长度的视频
				start = jglobal.Atoi[uint32](ab[0])
			}
		} else {
			// 分片的后续请求，返回指定片段的视频内容
			start = jglobal.Atoi[uint32](ab[0])
			end = jglobal.Atoi[uint32](ab[1])
		}
	}
	pack := &jglobal.Pack{
		Cmd: jpb.CMD_VIDEO_REQ,
		Data: &jpb.VideoReq{
			Uid:   jglobal.Atoi[uint32](parts[len(parts)-1]),
			Start: start,
			End:   end,
		},
	}
	han := jnet.Https.Handler[pack.Cmd]
	if han == nil {
		jlog.Errorf("no cmd(%s)", pack.Cmd)
		return
	}
	han.Fun(pack)
	if pack.Data == nil {
		return
	}
	rsp := pack.Data.(*jpb.VideoRsp)
	size := uint32(len(rsp.Video))
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Length", jglobal.Itoa(size))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, start+size-1, rsp.Size))
	w.WriteHeader(http.StatusPartialContent)
	jlog.Debug(w.Header())
	if _, err := w.Write(rsp.Video); err != nil {
		jlog.Error(err)
	}
}
