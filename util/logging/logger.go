package logging

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

func MakeLogger(prefix string) Logger {
	return log.New(os.Stderr, prefix, 0)
}
