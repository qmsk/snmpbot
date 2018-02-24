package server

import (
	"github.com/qmsk/go-logging"
)

var log logging.Logging

func SetLogging(l logging.Logging) {
	log = l
}
