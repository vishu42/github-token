package pkg

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// common errors.
var (
	ErrCantCreateLogger = errors.New("cant create logger")
)

// Logger is an interface for a logging client.
type Logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

// New returns new instance of logger.
func New(debug bool) (l *zap.SugaredLogger, err error) {
	var log *zap.Logger

	if debug {
		log, err = zap.NewDevelopment()
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.LevelKey = ""
		cfg.EncoderConfig.CallerKey = ""
		log, err = cfg.Build()
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCantCreateLogger, err)
	}

	return log.Sugar(), nil
}
