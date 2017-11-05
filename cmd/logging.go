package cmd

import (
	"log"
	"os"
)

func makeLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, 0)
}
