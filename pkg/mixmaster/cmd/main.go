package main

import (
	"flag"
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster"
)

var (
	configFile   = flag.String("config", "config.yaml", "config.yaml file to custom Mix Master")
	showSessions = flag.Bool("show-sessions", false, "Enable debug mode")
)

func main() {
	// Parse Flags
	flag.Parse()

	if *showSessions {
		fmt.Println("Sessions")
		return
	}

	mixmaster.NewMixMaster(configFile)

}
