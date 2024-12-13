package jlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var g_log *Log

type LogFormater struct{}

type Log struct {
	*logrus.Logger
	skip int
}

// ------------------------- inside -------------------------

func (format *LogFormater) Format(entry *logrus.Entry) ([]byte, error) {
	t := entry.Time.Format("2006-01-02 15:04:05.000")
	lv := strings.ToUpper(entry.Level.String())
	_, file, line, _ := runtime.Caller(g_log.skip)
	file = filepath.Base(file)
	return []byte(fmt.Sprintf("[%s][%s][%s.%d] %s\n", t, lv, file, line, entry.Message)), nil
}

func init() {
	g_log = &Log{}
	g_log.Logger = logrus.New()
	g_log.Out = os.Stdout
	g_log.SetLevel(logrus.TraceLevel)
	g_log.SetReportCaller(true)
	g_log.SetFormatter(&LogFormater{})
}

// ------------------------- outside -------------------------

func Logger() *Log {
	return g_log
}

func Trace(args ...any) {
	g_log.skip = 7
	g_log.Trace(args...)
}

func Tracef(format string, args ...any) {
	g_log.skip = 8
	g_log.Tracef(format, args...)
}

func Traceln(args ...any) {
	g_log.skip = 8
	g_log.Traceln(args...)
}

func Debug(args ...any) {
	g_log.skip = 7
	g_log.Debug(args...)
}

func Debugf(format string, args ...any) {
	g_log.skip = 8
	g_log.Debugf(format, args...)
}

func Debugln(args ...any) {
	g_log.skip = 8
	g_log.Debugln(args...)
}

func Info(args ...any) {
	g_log.skip = 7
	g_log.Info(args...)
}

func Infof(format string, args ...any) {
	g_log.skip = 8
	g_log.Infof(format, args...)
}

func Infoln(args ...any) {
	g_log.skip = 8
	g_log.Infoln(args...)
}

func Warn(args ...any) {
	g_log.skip = 7
	g_log.Warn(args...)
}

func Warnf(format string, args ...any) {
	g_log.skip = 8
	g_log.Warnf(format, args...)
}

func Warnln(args ...any) {
	g_log.skip = 8
	g_log.Warnln(args...)
}

func Error(args ...any) {
	g_log.skip = 7
	g_log.Error(args...)
}

func Errorf(format string, args ...any) {
	g_log.skip = 8
	g_log.Errorf(format, args...)
}

func Errorln(args ...any) {
	g_log.skip = 8
	g_log.Errorln(args...)
}

func Fatal(args ...any) {
	g_log.skip = 7
	g_log.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	g_log.skip = 8
	g_log.Fatalf(format, args...)
}

func Fatalln(args ...any) {
	g_log.skip = 8
	g_log.Fatalln(args...)
}

func Panic(args ...any) {
	g_log.skip = 7
	g_log.Panic(args...)
}

func Panicf(format string, args ...any) {
	g_log.skip = 8
	g_log.Panicf(format, args...)
}

func Panicln(args ...any) {
	g_log.skip = 8
	g_log.Panicln(args...)
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
