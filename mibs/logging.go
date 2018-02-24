package mibs

import (
	"github.com/qmsk/snmpbot/util/logging"
)

var log logging.Logging

func SetLogging(l logging.Logging) {
	log = l
}
