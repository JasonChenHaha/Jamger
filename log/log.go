package jlog

import (
	"fmt"
	"jconfig"
	"jglobal"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *Log

type LogFormater struct{}

type Log struct {
	*logrus.Logger
	skip int
}

// ------------------------- inside -------------------------

func (format *LogFormater) Format(entry *logrus.Entry) ([]byte, error) {
	t := entry.Time.Format("2006-01-02 15:04:05.000")
	lv := strings.ToUpper(entry.Level.String())
	_, file, line, _ := runtime.Caller(log.skip)
	file = filepath.Base(file)
	return []byte(fmt.Sprintf("[%s][%s][%s.%d] %s\n", t, lv, file, line, entry.Message)), nil
}

func Init() {
	log = &Log{}
	log.Logger = logrus.New()
	log.SetLevel(logrus.TraceLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&LogFormater{})
	output := &lumberjack.Logger{
		Filename:   "./log/" + jglobal.SERVER + ".log",
		MaxSize:    jconfig.GetInt("log.maxSize"),
		MaxBackups: jconfig.GetInt("log.maxBackup"),
		MaxAge:     jconfig.GetInt("log.maxAge"),
		Compress:   jconfig.GetBool("log.compress"),
	}
	log.SetOutput(output)
}

// ------------------------- outside -------------------------

func Logger() *Log {
	return log
}

func Trace(args ...any) {
	log.skip = 7
	log.Trace(args...)
}

func Tracef(format string, args ...any) {
	log.skip = 8
	log.Tracef(format, args...)
}

func Traceln(args ...any) {
	log.skip = 8
	log.Traceln(args...)
}

func Debug(args ...any) {
	log.skip = 7
	log.Debug(args...)
}

func Debugf(format string, args ...any) {
	log.skip = 8
	log.Debugf(format, args...)
}

func Debugln(args ...any) {
	log.skip = 8
	log.Debugln(args...)
}

func Info(args ...any) {
	log.skip = 7
	log.Info(args...)
}

func Infof(format string, args ...any) {
	log.skip = 8
	log.Infof(format, args...)
}

func Infoln(args ...any) {
	log.skip = 8
	log.Infoln(args...)
}

func Warn(args ...any) {
	log.skip = 7
	log.Warn(args...)
}

func Warnf(format string, args ...any) {
	log.skip = 8
	log.Warnf(format, args...)
}

func Warnln(args ...any) {
	log.skip = 8
	log.Warnln(args...)
}

func Error(args ...any) {
	log.skip = 7
	log.Error(args...)
}

func Errorf(format string, args ...any) {
	log.skip = 8
	log.Errorf(format, args...)
}

func Errorln(args ...any) {
	log.skip = 8
	log.Errorln(args...)
}

func Fatal(args ...any) {
	log.skip = 7
	log.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	log.skip = 8
	log.Fatalf(format, args...)
}

func Fatalln(args ...any) {
	log.skip = 8
	log.Fatalln(args...)
}

func Panic(args ...any) {
	log.skip = 7
	log.Panic(args...)
}

func Panicf(format string, args ...any) {
	log.skip = 8
	log.Panicf(format, args...)
}

func Panicln(args ...any) {
	log.skip = 8
	log.Panicln(args...)
}

// for nsq
func (log *Log) Output(calldepth int, s string) error {
	switch s[:3] {
	case "DEG":
		Debug(s[9:])
	case "INF":
		Info(s[9:])
	case "WRN":
		Warn(s[9:])
	case "ERR":
		Error(s[9:])
	}
	return nil
}
