package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Thoro/bfd-gobgp-connector/logging"
)

func main() {
	configPath := ""

	cmd := &cobra.Command{
		Use: "bfd-gobgp-connector",
		Run: func(_ *cobra.Command, _ []string) {
			runService(configPath)
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Execute()
}

func runService(configPath string) {
	config, err := LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	if len(config.Logging.Logfile) > 0 {
		log.SetLogfileName(config.Logging.Logfile)
	} else {
		log.SetLogfileName("interconnect.log")
	}
	log.SetLogToStdout(config.Logging.LogToStdout)

	service := NewInterconnectService(config)

	log.Infof("Starting server, quit using Ctrl+C")
	go service.Start()

	/* Make server killable by ^C */
	chanClose := make(chan os.Signal)
	signal.Notify(chanClose, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-chanClose:
		log.Infof("Shutting down server")
		return
	}
}
