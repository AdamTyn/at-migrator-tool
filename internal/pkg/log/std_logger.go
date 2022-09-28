package log

import (
	"fmt"
	"log"
)

var stdLogger = log.Default()

func __println(level string, vs ...interface{}) {
	vss := make([]interface{}, 0, len(vs)+1)
	vss = append(vss, level)
	vss = append(vss, vs...)
	stdLogger.Println(vss...)
}

func Debug(vs ...interface{}) {
	__println(DebugLevel, vs...)
}

func DebugF(format string, vs ...interface{}) {
	stdLogger.Println(DebugLevel, fmt.Sprintf(format, vs...))
}

func Info(vs ...interface{}) {
	__println(InfoLevel, vs...)
}

func InfoF(format string, vs ...interface{}) {
	stdLogger.Println(InfoLevel, fmt.Sprintf(format, vs...))
}

func Notice(vs ...interface{}) {
	__println(NoticeLevel, vs...)
}

func NoticeF(format string, vs ...interface{}) {
	stdLogger.Println(NoticeLevel, fmt.Sprintf(format, vs...))
}

func Warn(vs ...interface{}) {
	__println(WarnLevel, vs...)
}

func WarnF(format string, vs ...interface{}) {
	stdLogger.Println(WarnLevel, fmt.Sprintf(format, vs...))
}

func Exception(vs ...interface{}) {
	__println(ExceptionLevel, vs...)
}

func ExceptionF(format string, vs ...interface{}) {
	stdLogger.Println(ExceptionLevel, fmt.Sprintf(format, vs...))
}

func Error(vs ...interface{}) {
	__println(ErrorLevel, vs...)
}

func ErrorF(format string, vs ...interface{}) {
	stdLogger.Println(ErrorLevel, fmt.Sprintf(format, vs...))
}
