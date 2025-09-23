package main

import (
	"flag"
	"fmt"
	"os/exec"

	"fyne.io/systray"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/icon"
)

var (
	configFile   = flag.String("config", "../../../config.yaml", "config.yaml file to custom Mix Master")
	showSessions = flag.Bool("show-sessions", false, "Enable debug mode")
)

func main() {
	// Parse Flags
	flag.Parse()

	if *showSessions {
		fmt.Println("Sessions")
		return
	}

	systray.Run(onReady, onExit)

}

func onReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("MixMaster")
	config := systray.AddMenuItem("Config", "Configure Settings")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Stop deej and quit")

	go mixmaster.NewMixMaster(configFile)

	for {
		select {
		case <-quit.ClickedCh:
			systray.Quit()

		case <-config.ClickedCh:
			fmt.Println("Config Clicked")

			command := exec.Command("gedit", *configFile) // try gedit
			if err := command.Run(); err != nil {
				fmt.Println("Errors: ", err)
				command := exec.Command("kate", *configFile) // try kate instead
				if err := command.Run(); err != nil {
					fmt.Println("Errors: ", err)
				}
			}

		}
	}
}

func onExit() {
	fmt.Println("Exited")
}
