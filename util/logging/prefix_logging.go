package logging

func WithPrefix(logging Logging, prefix string) PrefixLogging {
	return PrefixLogging{logging, prefix}
}

type PrefixLogging struct {
	Logging Logging
	Prefix  string
}

func (pl PrefixLogging) WithPrefix(prefix string) PrefixLogging {
	return PrefixLogging{pl.Logging, pl.Prefix + ": " + prefix}
}

func (pl *PrefixLogging) Debugf(format string, args ...interface{}) {
	if pl != nil {
		pl.Logging.Debugf(pl.Prefix+": "+format, args...)
	}
}

func (pl *PrefixLogging) Infof(format string, args ...interface{}) {
	if pl != nil {
		pl.Logging.Infof(pl.Prefix+": "+format, args...)
	}
}

func (pl *PrefixLogging) Warnf(format string, args ...interface{}) {
	if pl != nil {
		pl.Logging.Warnf(pl.Prefix+": "+format, args...)
	}
}

func (pl *PrefixLogging) Errorf(format string, args ...interface{}) {
	if pl != nil {
		pl.Logging.Errorf(pl.Prefix+": "+format, args...)
	}
}
