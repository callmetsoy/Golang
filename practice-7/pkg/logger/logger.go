package logger

import "log"

type Interface interface {
	Info(msg string)
	Error(msg string)
}

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Info(msg string) {
	log.Println("INFO:", msg)
}

func (l *Logger) Error(msg string) {
	log.Println("ERROR:", msg)
}