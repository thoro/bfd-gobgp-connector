package main

import (
	"os"
	"os/signal"
	"syscall"

	"bitbucket.cf-it.at/creamfinance/gobgpd-bfdd-interconnect/logging"
)

func main() {
	// TODO: Enable different config files with -c parameter
	config, err := LoadConfig("")
	if err != nil {
		panic(err)
	}

	if len(config.Logging.Logfile) > 0 {
		log.SetLogfileName(config.Logging.Logfile)
	} else {
		log.SetLogfileName("interconnect.log")
	}
	log.SetLogToStdout(config.Logging.LogToStdout)

	log.Infof("Starting server, quit using Ctrl+C")

	service := NewInterconnectService(config)
	go service.Start()

	chanClose := make(chan os.Signal)
	signal.Notify(chanClose, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-chanClose:
		log.Infof("Shutting down server")
		return
	}
}
