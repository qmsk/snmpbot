package cmd

import (
	"github.com/qmsk/go-logging"
)

var Log logging.Logging // public for cmd packages

func SetLogging(l logging.Logging) {
	Log = l
}
