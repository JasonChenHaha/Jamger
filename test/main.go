package main

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jschedule"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

const (
	packSize     = 4
	uidSize      = 4
	cmdSize      = 2
	checksumSize = 4
	aesKeySize   = 16
)

var gateNum = 1
var aesKey []byte
var uid uint32
var id string
var pwd string
var httpAddr string
var httpsAddr string
var tcpAddr string

func makeAddr(addr string, key string) string {
	h := fnv.New32a()
	h.Write([]byte(key))
	n := h.Sum32() % uint32(gateNum)
	return fmt.Sprintf("%s%d", addr[0:len(addr)-1], n)
}

func main() {
	jconfig.Init()
	jglobal.Init()
	jlog.Init("")
	jschedule.Init()

	id, pwd = os.Args[2], os.Args[3]
	httpAddr = makeAddr(jconfig.GetString("http.addr"), id)
	httpsAddr = makeAddr(jconfig.GetString("https.addr"), id)
	tcpAddr = makeAddr(jconfig.GetString("tcp.addr"), id)
	var err error
	aesKey, err = jglobal.AesGenerate(16)
	if err != nil {
		jlog.Fatal(err)
	}
	testHttp()
	// testHttps()
	testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
}

func LoadImage(name string) (string, []byte) {
	file, err := os.Open(name)
	if err != nil {
		jlog.Error(err)
		return "", nil
	}
	defer file.Close()
	_, format, err := image.Decode(file)
	if err != nil {
		jlog.Error(err)
		return "", nil
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		jlog.Error(err)
		return "", nil
	}
	img, err := io.ReadAll(file)
	if err != nil {
		jlog.Error(err)
		return "", nil
	}
	return format, img
}

func CompressJPG(file *os.File, dir, name, ext string) {
	img, err := jpeg.Decode(file)
	if err != nil {
		jlog.Error("无法解码图片:", err)
		return
	}
	// 调整尺寸
	img = imaging.Resize(img, 250, 0, imaging.Lanczos)
	out, err := os.Create(fmt.Sprintf("%s/%s_s%s", dir, name, ext))
	if err != nil {
		jlog.Error("无法创建输出文件:", err)
		return
	}
	defer out.Close()
	if err = jpeg.Encode(out, img, &jpeg.Options{Quality: 50}); err != nil {
		jlog.Error("无法写入输出文件:", err)
		return
	}
}

func CompressPNG(file *os.File, dir, name, ext string) {
	img, err := png.Decode(file)
	if err != nil {
		jlog.Error("无法解码图片:", err)
		return
	}
	// 调整尺寸
	img = imaging.Resize(img, 250, 0, imaging.Lanczos)
	out, err := os.Create(fmt.Sprintf("%s/%s_s%s", dir, name, ext))
	if err != nil {
		jlog.Error("无法创建输出文件:", err)
		return
	}
	defer out.Close()
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err = encoder.Encode(out, img); err != nil {
		jlog.Error("无法写入输出文件:", err)
		return
	}
}

func CompressImage(name string) {
	dir := filepath.Dir(name)
	fileName := filepath.Base(name)
	ext := filepath.Ext(name)
	fileName = fileName[:len(fileName)-len(ext)]
	file, err := os.Open(name)
	if err != nil {
		jlog.Error(err)
		return
	}
	defer file.Close()
	// 图像大于阈值才压缩
	switch ext {
	case ".png":
		CompressPNG(file, dir, fileName, ext)
	case ".jpg", ".jpeg":
		CompressJPG(file, dir, fileName, ext)
	}
}
