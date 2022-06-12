package logger

import (
	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	logr.Logger
	logrusLog *logrus.Logger
}

const defaultLevel = logrus.ErrorLevel

func (l *Logger) SetLevel(levelText string) {
	var level logrus.Level
	var err error
	if level, err = logrus.ParseLevel(levelText); err != nil {
		level = defaultLevel
	}

	l.logrusLog.SetLevel(level)
}

func New() *Logger {
	logrusLog := logrus.New()
	logrusLog.SetLevel(defaultLevel)

	return &Logger{
		Logger:    logrusr.New(logrusLog),
		logrusLog: logrusLog,
	}
}

func MapToKV(m map[string]interface{}) []interface{} {
	pairs := []interface{}{}
	for k, v := range m {
		pairs = append(pairs, k, v)
	}

	return pairs
}
