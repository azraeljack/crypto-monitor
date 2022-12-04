package main

import (
	"context"
	"flag"
	"github.com/azraeljack/crypto-monitor/logging"
	_ "github.com/azraeljack/crypto-monitor/logging"
	"github.com/azraeljack/crypto-monitor/monitor"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	config := flag.String("config", "./config.json", "monitor config json file path")
	flag.Parse()

	if !*debug {
		logging.SetupLogRotate()
	} else {
		log.SetLevel(log.DebugLevel)
		log.Debug("debug mode enabled")
	}

	log.Info("starting the monitor...")

	ctx, cancel := context.WithCancel(context.Background())

	mon := monitor.NewMonitor(ctx, *config)
	mon.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT)

	<-c
	<-mon.Stop()

	cancel()

	log.Info("monitor exited")
}
