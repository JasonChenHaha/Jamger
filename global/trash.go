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
	"jconfig"
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

func GetGroup(cmd jpb.CMD) int {
	return int(cmd / jpb.CMD_MAX)
}

// 序列化服务器基本信息
func SerializeServerInfo() string {
	info := map[string]any{"addr": jconfig.GetString("rpc.addr")}
	data, err := json.Marshal(info)
	if err != nil {
		jlog.Panic(err)
	}
	return string(data)
}

// 反序列化服务器基本信息
func UnserializeServerInfo(data []byte) (info map[string]any) {
	if err := json.Unmarshal(data, &info); err != nil {
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

// RSA生成钥匙对
func RSAGenerate() (string, string, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	pub := &pri.PublicKey
	priBytes := x509.MarshalPKCS1PrivateKey(pri)
	pubBytes := x509.MarshalPKCS1PublicKey(pub)
	priP := &pem.Block{Bytes: priBytes}
	pubP := &pem.Block{Bytes: pubBytes}
	return string(pem.EncodeToMemory(priP)), string(pem.EncodeToMemory(pubP)), nil
}

// RSA加载公钥
func RSALoadPublicKey(publicKey string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, fmt.Errorf("decode publicKey failed.")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

// RSA加载私钥
func RSALoadPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, fmt.Errorf("decode privateKey failed.")
	}
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pri, nil
}

// RSA公钥加密
func RSAEncrypt(pubKey *rsa.PublicKey, data *[]byte) (err error) {
	*data, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, *data, nil)
	return
}

// RSA私钥解密
func RSADecrypt(privKey *rsa.PrivateKey, data *[]byte) (err error) {
	*data, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, *data, nil)
	return
}

// AES生成密钥
func AESGenerate(size int) ([]byte, error) {
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// AES加密
func AESEncrypt(key []byte, data *[]byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	blockSize := block.BlockSize()
	padding := blockSize - len(*data)%blockSize
	*data = append(*data, bytes.Repeat([]byte{byte(padding)}, padding)...)
	secret := make([]byte, blockSize+len(*data))
	iv := secret[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(secret[blockSize:], *data)
	*data = secret
	return nil
}

// AES解密
func AESDecrypt(key []byte, data *[]byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	blockSize := block.BlockSize()
	if len(*data) < blockSize {
		return fmt.Errorf("data too short.")
	}
	iv := (*data)[:blockSize]
	*data = (*data)[blockSize:]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(*data, *data)
	*data = (*data)[:len(*data)-int((*data)[len(*data)-1])]
	return nil
}

// 生成token
func TokenGenerate(base string) (string, error) {
	a := []byte(base)
	size := len(a)
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	c := time.Now().Unix()
	raw := make([]byte, size+72)
	copy(raw, a)
	copy(raw[size:], b)
	binary.NativeEndian.PutUint64(raw[size+3:], uint64(c))
	hash := md5.New()
	hash.Write(raw)
	hashStr := hex.EncodeToString(hash.Sum(nil))
	return hashStr, nil
}

func Atoi[T AllInt](data string) T {
	n, err := strconv.Atoi(data)
	if err != nil {
		jlog.Panic(err)
	}
	return T(n)
}
