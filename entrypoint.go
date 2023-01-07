package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

var globalruntimeparams runtimeparams
var voiceSessionMaster []voicesessions

func main() {

	initConfigs()
	go discordModule()
	go httpModule()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	defer closeAllConnections()

}
