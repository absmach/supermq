package influxdb

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
)

type loggingMiddleware struct {
	logger log.Logger
	w      Writer
}

// LoggingMiddleware adds logging facilities to the core writer.
func LoggingMiddleware(w Writer, logger log.Logger) Writer {
	return &loggingMiddleware{logger, w}
}

func (lm *loggingMiddleware) Save(msg mainflux.Message) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method save took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.w.Save(msg)
}

func (lm *loggingMiddleware) Close() (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method close took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.w.Close()
}
