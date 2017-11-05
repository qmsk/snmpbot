package client

type Logger interface {
	Printf(format string, args ...interface{})
}

type Logging struct {
	Debug Logger
	Info  Logger
	Warn  Logger
	Error Logger
}

func (log *Logging) Debugf(format string, args ...interface{}) {
	if log != nil && log.Debug != nil {
		log.Debug.Printf(format, args...)
	}
}

func (log *Logging) Infof(format string, args ...interface{}) {
	if log != nil && log.Info != nil {
		log.Info.Printf(format, args...)
	}
}

func (log *Logging) Warnf(format string, args ...interface{}) {
	if log != nil && log.Warn != nil {
		log.Warn.Printf(format, args...)
	}
}

func (log *Logging) Errorf(format string, args ...interface{}) {
	if log != nil && log.Error != nil {
		log.Error.Printf(format, args...)
	}
}
