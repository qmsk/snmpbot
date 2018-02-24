package logging

type Logging struct {
	Debug Logger
	Info  Logger
	Warn  Logger
	Error Logger
}

func (logging *Logging) Debugf(format string, args ...interface{}) {
	if logging != nil && logging.Debug != nil {
		logging.Debug.Printf(format, args...)
	}
}

func (logging *Logging) Infof(format string, args ...interface{}) {
	if logging != nil && logging.Info != nil {
		logging.Info.Printf(format, args...)
	}
}

func (logging *Logging) Warnf(format string, args ...interface{}) {
	if logging != nil && logging.Warn != nil {
		logging.Warn.Printf(format, args...)
	}
}

func (logging *Logging) Errorf(format string, args ...interface{}) {
	if logging != nil && logging.Error != nil {
		logging.Error.Printf(format, args...)
	}
}
