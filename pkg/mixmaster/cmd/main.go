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
		fmt.Println(mixmaster.ListDevices())
		return
	}

	systray.Run(onReady, onExit)

}

func onReady() {
	var dat *mixmaster.ParsedAudioData
	// Create Pulse Client
	pulseClient, err := mixmaster.CreatePulseClient("MixMaster")
	if err != nil {
		//return errors.New("could not creating Pulse client")
		return
	}

	// Set up mpris
	mprisClient, err := mixmaster.MprisInitialize()
	if err != nil {
		//return errors.New("could not creating mpirs client")
		return
	}

	// Systray setup
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("MixMaster")
	config := systray.AddMenuItem("Config", "Configure Settings")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Stop deej and quit")

	// USB HID audio Device
	cfg1 := mixmaster.ParseConfig("../../../myconfig.yaml") //*configFile)
	dev, _ := mixmaster.GetDevice(6455666)
	mixmaster1out := make(chan *mixmaster.ParsedAudioData)
	go mixmaster.NewMixMaster(cfg1, dev, mixmaster1out)

	for {
		dat = <-mixmaster1out

		//fmt.Print("mixMaster1 data: ")
		//fmt.Println(dat)

		// Get pulse audio sessions
		pulseSessions, err := pulseClient.GetPulseSessions()
		if err != nil {
			// could not get pulse audio sessions
			return
		}

		// Get mpris sessions
		mpirsSessions, err := mprisClient.ConnectToApps(pulseSessions)
		if err != nil {
			// could not get app media controls
			return
		}

		pulseSessions.ChangeAppVolume(dat.PulseApps, pulseClient)
		pulseSessions.ChangeMasterVolume(dat.MasterOuputs, pulseClient)
		if false {
			fmt.Println(mpirsSessions)
		}
		//mpirsSessions.MediaControlApps(dat.MpirsApps, mpirsSessions)

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
		default:
			// no option selected
		}
	}
}

func onExit() {
	fmt.Println("Exited")
}
