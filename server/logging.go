package server

import (
	"github.com/qmsk/snmpbot/util/logging"
)

var log logging.Logging

func SetLogging(l logging.Logging) {
	log = l
}

func ApplyLogging(options logging.Options) {
	log = options.MakeLogging()
}
