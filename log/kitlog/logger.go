package kitlog

import (
	"io"

	logkit "github.com/go-kit/kit/log"
	"github.com/mainflux/mainflux/log"
)

var _ log.Logger = (*logger)(nil)

type logger struct {
	lvl       log.Level
	kitLogger logkit.Logger
}

// New returns wrapped go kit logger.
func New(out io.Writer) log.Logger {
	l := logkit.NewJSONLogger(logkit.NewSyncWriter(out))
	l = logkit.With(l, "ts", logkit.DefaultTimestampUTC)
	return &logger{log.All, l}
}

func (l *logger) SetLevel(lvl log.Level) {
	l.lvl = lvl
}

func (l logger) Info(obj interface{}) {
	l.Log(log.Info, obj)
}

func (l logger) Warn(obj interface{}) {
	l.Log(log.Warn, obj)
}

func (l logger) Error(obj interface{}) {
	l.Log(log.Error, obj)
}

func (l logger) Log(lvl log.Level, obj interface{}) {
	if lvl >= l.lvl {
		l.kitLogger.Log(lvl.String(), obj)
	}
}
