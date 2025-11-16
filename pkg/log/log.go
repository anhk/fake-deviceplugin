package log

import (
	"context"
	"fake-deviceplugin/pkg/utils"
	"fmt"
	"log"
	"os"
)

type LEVEL int

const (
	DEBUG LEVEL = iota
	INFO
	WARN
	ERROR
	NONE
)

var (
	logger *log.Logger
	level  = DEBUG
)

func init() {
	logger = log.New(os.Stdout, "", 0)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func SetLogger(l *log.Logger) {
	logger = l
}

func SetLevel(l LEVEL) {
	level = l
}

func logPrint(prefix string, v ...any) {
	var arr []interface{}
	arr = append(arr, prefix)
	arr = append(arr, v...)
	_ = logger.Output(3, fmt.Sprintln(arr...))
}

func traceId(ctx context.Context) string {
	traceId := utils.GetTraceId(ctx)
	return fmt.Sprintf("[%s]", traceId)
}

func Debug(ctx context.Context, v ...any) {
	if level <= DEBUG {
		logPrint("[DEBUG]"+traceId(ctx), v...)
	}
}

func Debugf(ctx context.Context, f string, v ...interface{}) {
	if level <= DEBUG {
		logPrint("[DEBUG]"+traceId(ctx), fmt.Sprintf(f, v...))
	}
}

func Info(ctx context.Context, v ...any) {
	if level <= DEBUG {
		logPrint("[INFO]"+traceId(ctx), v...)
	}
}

func Infof(ctx context.Context, f string, v ...interface{}) {
	if level <= DEBUG {
		logPrint("[INFO]"+traceId(ctx), fmt.Sprintf(f, v...))
	}
}

func Warn(ctx context.Context, v ...any) {
	if level <= DEBUG {
		logPrint("[WARN]"+traceId(ctx), v...)
	}
}

func Warnf(ctx context.Context, f string, v ...interface{}) {
	if level <= DEBUG {
		logPrint("[WARN]"+traceId(ctx), fmt.Sprintf(f, v...))
	}
}

func Error(ctx context.Context, v ...any) {
	if level <= DEBUG {
		logPrint("[ERROR]"+traceId(ctx), v...)
	}
}

func Errorf(ctx context.Context, f string, v ...interface{}) {
	if level <= DEBUG {
		logPrint("[ERROR]"+traceId(ctx), fmt.Sprintf(f, v...))
	}
}
