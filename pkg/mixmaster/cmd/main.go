package main

import (
	"flag"
	"fmt"
	"os/exec"
	"time"

	"fyne.io/systray"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/icon"
	"github.com/tarm/serial"
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
		fmt.Println(mixmaster.ListDevices())
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

	cfg1 := mixmaster.ParseConfig("../../../myconfig.yaml") //*configFile)
	dev, _ := mixmaster.GetDevice(6455666)
	go mixmaster.NewMixMaster(cfg1, dev)

	cfg2 := mixmaster.ParseConfig("../../../config.yaml")
	c := &serial.Config{
		Name:        cfg2.COMPort,
		Baud:        cfg2.BaudRate,
		ReadTimeout: time.Second * 5,
	}
	go mixmaster.NewMixMaster(cfg2, &mixmaster.Device{HidDev: nil, SerialDev: c})

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
