package main

import (
	"flag"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster"
)

var (
	configFile   = flag.String("config", "../../../config.yaml", "config.yaml file to custom Mix Master")
	showSessions = flag.Bool("show-sessions", false, "Enable debug mode")
)

func main() {
	// Parse Flags
	flag.Parse()

	if *showSessions {
		dev, _ := mixmaster.GetDevice()
		fmt.Println("Device List", dev)
		return
	}

	a := app.New()
	a.SendNotification(fyne.NewNotification("Mixmaster", "App has been started"))
	w := a.NewWindow("MixMaster")

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("MixMaster",
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}
	w.SetContent(widget.NewLabel("Mixmaster"))
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	go func() {
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

		// USB HID audio Device
	test:
		dev, _ := mixmaster.GetDevice()
		cfg := mixmaster.ParseConfig("../../../myconfig.yaml") //*configFile)
		deviceTest, err := mixmaster.NewMixMaster(cfg, dev.HidDev[10051537])
		if err != nil {
			fmt.Println("Failed to Find Device")
			time.Sleep(5 * time.Second)
			goto test
		}

		for {
			dat, err := deviceTest.Pull(cfg)
			if err != nil {
				fmt.Println("Error Pulling Data:", err)
			repeat:
				time.Sleep(5 * time.Second)
				dev, _ := mixmaster.GetDevice()
				deviceTest, _ = mixmaster.NewMixMaster(cfg, dev.HidDev[10051537])
				if deviceTest == nil {
					fmt.Println("Did not find device:", dev.HidDev[10051537])
					goto repeat
				}
				fmt.Println("Found")
				continue
			}

			// Get pulse audio sessions
			pulseSessions, err := pulseClient.GetPulseSessions()
			if err != nil {
				// could not get pulse audio sessions
				return
			}

			// Get mpris sessions
			mpirsSessions, err := mprisClient.GetMpirsSessions()
			if err != nil {
				// could not get app media controls
				return
			}
			pulseSessions.ChangeAppVolume(dat.PulseApps, pulseClient)
			pulseSessions.ChangeMasterVolume(dat.MasterOuputs, pulseClient)
			mpirsSessions.MediaControls(dat.MpirsApps, mprisClient)
		}
	}()

	a.Run()
}
