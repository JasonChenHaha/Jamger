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
	log  *logrus.Logger
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
	g_log = &Log{log: logrus.New()}
	g_log.log.Out = os.Stdout
	g_log.log.SetLevel(logrus.TraceLevel)
	g_log.log.SetReportCaller(true)
	g_log.log.SetFormatter(&LogFormater{})
}

// ------------------------- outside -------------------------

func Logger() *logrus.Logger {
	return g_log.log
}

func Trace(args ...any) {
	g_log.skip = 7
	g_log.log.Trace(args...)
}

func Tracef(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Tracef(format, args...)
}

func Traceln(args ...any) {
	g_log.skip = 8
	g_log.log.Traceln(args...)
}

func Debug(args ...any) {
	g_log.skip = 7
	g_log.log.Debug(args...)
}

func Debugf(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Debugf(format, args...)
}

func Debugln(args ...any) {
	g_log.skip = 8
	g_log.log.Debugln(args...)
}

func Info(args ...any) {
	g_log.skip = 7
	g_log.log.Info(args...)
}

func Infof(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Infof(format, args...)
}

func Infoln(args ...any) {
	g_log.skip = 8
	g_log.log.Infoln(args...)
}

func Warn(args ...any) {
	g_log.skip = 7
	g_log.log.Warn(args...)
}

func Warnf(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Warnf(format, args...)
}

func Warnln(args ...any) {
	g_log.skip = 8
	g_log.log.Warnln(args...)
}

func Error(args ...any) {
	g_log.skip = 7
	g_log.log.Error(args...)
}

func Errorf(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Errorf(format, args...)
}

func Errorln(args ...any) {
	g_log.skip = 8
	g_log.log.Errorln(args...)
}

func Fatal(args ...any) {
	g_log.skip = 7
	g_log.log.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Fatalf(format, args...)
}

func Fatalln(args ...any) {
	g_log.skip = 8
	g_log.log.Fatalln(args...)
}

func Panic(args ...any) {
	g_log.skip = 7
	g_log.log.Panic(args...)
}

func Panicf(format string, args ...any) {
	g_log.skip = 8
	g_log.log.Panicf(format, args...)
}

func Panicln(args ...any) {
	g_log.skip = 8
	g_log.log.Panicln(args...)
}
