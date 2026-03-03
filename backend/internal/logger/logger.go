package logger

import (
	"log"
	"os"
)

var (
	infoLog  = log.New(os.Stdout, "", log.LstdFlags)
	errorLog = log.New(os.Stderr, "", log.LstdFlags)
)

func Info(format string, args ...any) {
	infoLog.Printf(format, args...)
}

func Infoln(msg string) {
	infoLog.Println(msg)
}

func Error(format string, args ...any) {
	errorLog.Printf(format, args...)
}

func Fatal(format string, args ...any) {
	errorLog.Fatalf(format, args...)
}
