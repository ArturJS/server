package main

import (
	"github.com/getsentry/sentry-go"
	l "github.com/loeffel-io/logger"
	"github.com/mholt/archiver/v3"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

func main() {
	var (
		err error
		zip = new(archiver.Zip)
	)

	// Setup sentry
	if err = sentry.Init(sentry.ClientOptions{Dsn: os.Getenv("SENTRY")}); err != nil {
		log.Fatal(err)
	}

	// Logger
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	logger := &l.Logger{
		Debug:   true,
		RWMutex: new(sync.RWMutex),
	}

	// api
	api := &api{
		zip:     zip,
		port:    "8080",
		mode:    "debug",
		RWMutex: new(sync.RWMutex),
	}

	if err = api.startServer(); err != nil {
		logger.Error(err)
	}
}
