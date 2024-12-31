package jtrash

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"jconfig"
	"jlog"
	"os"
	"os/signal"
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
	block, _ := pem.Decode([]byte(jconfig.GetString("rsa.privateKey")))
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
func RSAEncrypt(pubKey *rsa.PublicKey, data []byte) ([]byte, error) {
	secret, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, data, nil)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// RSA私钥解密
func RSADecrypt(privKey *rsa.PrivateKey, secret []byte) ([]byte, error) {
	unsecret, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, secret, nil)
	if err != nil {
		return nil, err
	}
	return unsecret, nil
}
