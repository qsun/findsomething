package main

import "os/signal"
import "log"
import "os"
import "flag"

var monitoringDir = flag.String("dir", ".", "Directory to be monitored")
var searchSock = flag.String("search", "search.sock", "sock file for search")

func main() {
	log.Println("Started.")
	flag.Parse()

	// defer os.Remove(*searchSock)

	monitor := NewMonitoring(*monitoringDir, *searchSock)
	go monitor.Start()
	go monitor.StartSearch()

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, os.Interrupt)

	for {
		select {
		case e := <-monitor.Change:
			monitor.ProcessEvent(e)
		case e := <-signal_chan:
			log.Println("Got event", e)
			log.Println("Clear")
			os.Remove(*searchSock)
			os.Exit(1)
		}
	}
}
