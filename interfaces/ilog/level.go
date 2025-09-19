package ilog

import (
	"strings"
)

type Level uint8

const (
	Debug Level = iota + 1
	Info
	Warn
	Error
)

var levelMapping = map[string]Level{
	"debug":   Debug,
	"info":    Info,
	"warn":    Warn,
	"warning": Warn,
	"error":   Error,
}

var levelString = map[Level]string{
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
}

func (level Level) String() string {
	return levelString[level]
}

func (level Level) EnableAll() bool {
	return level <= Debug
}

func (level Level) DisableAll() bool {
	return level > Error
}

func ParseLevel(in string) Level {
	l, ok := levelMapping[strings.ToLower(in)]
	if !ok {
		l = Info
	}
	return l
}
