package main

import (
	"flag"
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/icon"
	"github.com/getlantern/systray"
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

	systray.Run(onReady, onExit)

}

func onReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("MixMaster")
	config := systray.AddMenuItem("Config", "Configure Settings")
	systray.AddSeparator()
	check := systray.AddMenuItemCheckbox("Check", "Please Check Me", false)
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Stop deej and quit")

	go mixmaster.NewMixMaster(configFile)

	for {
		select {
		case <-quit.ClickedCh:
			systray.Quit()

		case <-config.ClickedCh:
			fmt.Println("Config Clicked")

		case <-check.ClickedCh:
			if check.Checked() {
				check.Uncheck()
				fmt.Println("Unchecked")
			} else {
				check.Check()
				fmt.Println("Checked")
			}
		}
	}
}

func onExit() {
	fmt.Println("Exit")
}
