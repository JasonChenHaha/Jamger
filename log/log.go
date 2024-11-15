package jlog

import (
	"os"

	"github.com/sirupsen/logrus"
)

var g_log *logrus.Logger

func init() {
	g_log = logrus.New()
	g_log.Out = os.Stdout
	g_log.SetLevel(logrus.TraceLevel)
	g_log.SetReportCaller(true)
}

func Trace(args ...any) {
	g_log.Trace(args...)
}

func Tracef(format string, args ...any) {
	g_log.Tracef(format, args...)
}

func Traceln(args ...any) {
	g_log.Traceln(args...)
}

func Debug(args ...any) {
	g_log.Debug(args...)
}

func Debugf(format string, args ...any) {
	g_log.Debugf(format, args...)
}

func Debugln(args ...any) {
	g_log.Debugln(args...)
}

func Info(args ...any) {
	g_log.Info(args...)
}

func Infof(format string, args ...any) {
	g_log.Infof(format, args...)
}

func Infoln(args ...any) {
	g_log.Infoln(args...)
}

func Warn(args ...any) {
	g_log.Warn(args...)
}

func Warnf(format string, args ...any) {
	g_log.Warnf(format, args...)
}

func Warnln(args ...any) {
	g_log.Warnln(args...)
}

func Error(args ...any) {
	g_log.Error(args...)
}

func Errorf(format string, args ...any) {
	g_log.Errorf(format, args...)
}

func Errorln(args ...any) {
	g_log.Errorln(args...)
}

func Fatal(args ...any) {
	g_log.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	g_log.Fatalf(format, args...)
}

func Fatalln(args ...any) {
	g_log.Fatalln(args...)
}

func Panic(args ...any) {
	g_log.Panic(args...)
}

func Panicf(format string, args ...any) {
	g_log.Panicf(format, args...)
}

func Panicln(args ...any) {
	g_log.Panicln(args...)
}
