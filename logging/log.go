package logging

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func SetupLogRotate() {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.ToSlash("/root/monitor.log"),
		MaxSize:    500, // MB
		MaxBackups: 3,
		MaxAge:     30, // days
		Compress:   true,
	}

	// Fork writing into two outputs
	multiWriter := io.MultiWriter(os.Stderr, lumberjackLogger)

	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = time.RFC1123Z // or RFC3339
	logFormatter.FullTimestamp = true

	log.SetFormatter(logFormatter)
	log.SetLevel(log.InfoLevel)
	log.SetOutput(multiWriter)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			if err := lumberjackLogger.Rotate(); err != nil {
				return
			}
		}
	}()
}
