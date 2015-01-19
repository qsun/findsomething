package main

import "log"
import "flag"

var monitoringDir = flag.String("dir", ".", "Directory to be monitored")

func main() {
	log.Println("Started.")
	flag.Parse()

	monitor := NewMonitoring(*monitoringDir)
	go monitor.Start()

	for {
		e := <-monitor.Change
		monitor.ProcessEvent(e)
	}
}
