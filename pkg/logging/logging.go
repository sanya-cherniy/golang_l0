package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		w.Write([]byte(line))
	}
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry

var entryes = make(map[string]*logrus.Entry)

type Logger struct {
	*logrus.Entry
}

func GetLogger(file string) (*Logger, error) {
	e, ok := entryes[file]
	if !ok {
		return nil, errors.New(fmt.Sprintf("log file %s not found", file))
	}
	return &Logger{e}, nil
}

func (l *Logger) GetLoggerWithField(k string, v interface{}) *Logger {
	return &Logger{l.WithField(k, v)}
}

func init() {
	err := os.MkdirAll("logs", 0766)
	if err != nil {
		panic(err)
	}
}

func InitLogger(logFile string) {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.JSONFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			return fmt.Sprintf("%s()", frame.Function), fmt.Sprintf("%s:%d", filename, frame.Line)
		},
	}

	allFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		panic(err)
	}

	l.SetOutput(io.Discard)

	l.AddHook(&writerHook{
		Writer:    []io.Writer{allFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	l.SetLevel(logrus.TraceLevel)
	entryes[logFile] = logrus.NewEntry(l)
}
