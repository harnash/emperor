package server

import (
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-playground/lars"
)

// LoggingAndRecovery handle HTTP request logging + recovery
func LoggingAndRecovery(c lars.Context) {

	t1 := time.Now()

	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1<<16)
			n := runtime.Stack(trace, true)
			log.WithFields(log.Fields{
				"error": err,
				"stack": trace[:n],
			}).Error("Recovering from panic")
			HandlePanic(c, trace[:n])
			return
		}
	}()

	c.Next()

	res := c.Response()
	req := c.Request()
	code := res.Status()

	t2 := time.Now()

	log.WithFields(log.Fields{
		"method": req.Method,
		"code":   code,
		"url":    req.URL,
		"time":   t2.Sub(t1),
		"size":   res.Size(),
	}).Info("Request")
}

// HandlePanic handles graceful panic by redirecting to friendly error page or rendering a friendly error page.
// trace passed just in case you want rendered to developer when not running in production
func HandlePanic(c lars.Context, trace []byte) {

	// redirect to or directly render friendly error page
}
