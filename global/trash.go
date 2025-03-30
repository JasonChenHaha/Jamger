package jglobal

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"jlog"
	"jpb"
	"os"
	"os/signal"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

// ------------------------- outside -------------------------

func Rcover() {
	if r := recover(); r != nil {
		jlog.Panic(r)
	}
}

// 阻塞进程
func Keep() {
	mainC := make(chan os.Signal, 1)
	signal.Notify(mainC, os.Interrupt)
	<-mainC
}

// 从cmd获取group
func GetGroup(cmd jpb.CMD) int {
	return int(cmd / jpb.CMD_MAX)
}

// 序列化服务器基本信息
func SerializeJson(data any) []byte {
	jData, err := json.Marshal(data)
	if err != nil {
		jlog.Panic(err)
	}
	return jData
}

// 反序列化服务器基本信息，data需要为指针类型
func UnserializeJson(jData []byte, data any) {
	if err := json.Unmarshal(jData, data); err != nil {
		jlog.Panic(err)
	}
	return
}

// grpc拦截器
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// 今天0点的时刻
func GetTodayZeroTime() *time.Time {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return &zero
}

// 明天0点的时间
func GetTomorrowZeroTime() *time.Time {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return &zero
}

// 到下秒的时间差
func TimeToSecond() time.Duration {
	now := time.Now()
	return time.Second - time.Duration(now.Nanosecond())
}

// 到下分钟的时间差
func TimeToMinute() time.Duration {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	return zero.Sub(now)
}

// 到明天0点的时间差
func TimeToTomorrow() time.Duration {
	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return zero.Sub(now)
}

// 到指定时刻的时间差
func TimeToTime(hour int) time.Duration {
	now := time.Now()
	time5 := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if now.After(time5) {
		time5 = time5.Add(24 * time.Hour)
	}
	return time5.Sub(now)
}

// Rsa生成钥匙对
func RsaGenerate() (string, string, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		jlog.Error(err)
		return "", "", err
	}
	pub := &pri.PublicKey
	priBytes := x509.MarshalPKCS1PrivateKey(pri)
	pubBytes := x509.MarshalPKCS1PublicKey(pub)
	priP := &pem.Block{Bytes: priBytes}
	pubP := &pem.Block{Bytes: pubBytes}
	return string(pem.EncodeToMemory(priP)), string(pem.EncodeToMemory(pubP)), nil
}

// Rsa加载公钥
func RsaLoadPublicKey(publicKey string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		err := fmt.Errorf("decode publicKey failed.")
		jlog.Error(err)
		return nil, err
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	return pub, nil
}

// Rsa加载私钥
func RsaLoadPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		err := fmt.Errorf("decode privateKey failed.")
		jlog.Error(err)
		return nil, err
	}
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		jlog.Error(err)
		return nil, err
	}
	return pri, nil
}

// Rsa公钥加密
func RsaEncrypt(pubKey *rsa.PublicKey, data *[]byte) (err error) {
	*data, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, *data, nil)
	if err != nil {
		jlog.Error(err)
	}
	return
}

// Rsa私钥解密
func RsaDecrypt(privKey *rsa.PrivateKey, data *[]byte) (err error) {
	*data, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, *data, nil)
	if err != nil {
		jlog.Error(err)
	}
	return
}

// Aes生成密钥
func AesGenerate(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	if err != nil {
		jlog.Error(err)
	}
	return key, err
}

// Aes加密
func AesEncrypt(key []byte, data *[]byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		jlog.Error(err)
		return err
	}
	blockSize := block.BlockSize()
	padding := blockSize - len(*data)%blockSize
	*data = append(*data, bytes.Repeat([]byte{byte(padding)}, padding)...)
	secret := make([]byte, blockSize+len(*data))
	iv := secret[:blockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		jlog.Error(err)
		return err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(secret[blockSize:], *data)
	*data = secret
	return err
}

// Aes解密
func AesDecrypt(key []byte, data *[]byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		jlog.Error(err)
		return err
	}
	blockSize := block.BlockSize()
	if len(*data) < blockSize {
		err = fmt.Errorf("data too short.")
		jlog.Error(err)
		return err
	}
	iv := (*data)[:blockSize]
	*data = (*data)[blockSize:]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(*data, *data)
	size := len(*data)
	if size == 0 {
		err = fmt.Errorf("data too short.")
		jlog.Error(err)
		return err
	}
	pos := size - int((*data)[size-1])
	if pos < 0 {
		err = fmt.Errorf("aes decrypt failed.")
		jlog.Error(err)
		return err
	}
	*data = (*data)[:pos]
	return nil
}

// 生成token
func TokenGenerate(base string) (string, error) {
	a := []byte(base)
	size := len(a)
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		jlog.Error(err)
		return "", err
	}
	c := time.Now().Unix()
	raw := make([]byte, size+72)
	copy(raw, a)
	copy(raw[size:], b)
	binary.NativeEndian.PutUint64(raw[size+3:], uint64(c))
	hash := md5.New()
	hash.Write(raw)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// string -> number
func Atoi[T AllInt](data string) T {
	n, err := strconv.Atoi(data)
	if err != nil {
		jlog.Panic(err)
	}
	return T(n)
}

// number -> string
func Itoa(data any) string {
	switch o := data.(type) {
	case int:
		return strconv.Itoa(o)
	case int32:
		return strconv.Itoa(int(o))
	case int64:
		return strconv.FormatInt(o, 10)
	case uint:
		return strconv.FormatUint(uint64(o), 10)
	case uint32:
		return strconv.FormatUint(uint64(o), 10)
	case uint64:
		return strconv.FormatUint(o, 10)
	case float32:
		return fmt.Sprintf("%f", o)
	case float64:
		return fmt.Sprintf("%f", o)
	}
	return ""
}

// url转成cmd
func UrlToCmd(str string) jpb.CMD {
	s := str[1:]
	if s == "" {
		return jpb.CMD_NIL
	} else {
		return Atoi[jpb.CMD](s)
	}
}

func Max[T AllInt](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func Min[T AllInt](a, b T) T {
	if a < b {
		return a
	}
	return b
}
