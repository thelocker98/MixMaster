package main

import (
	"flag"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster"
	ui "gitea.locker98.com/locker98/Mixmaster/pkg/ui"
)

// Setup CLI Flags and Default Values
var (
	configFile   = flag.String("config", "../../../config.yaml", "config.yaml file to custom Mix Master")
	showSessions = flag.Bool("show-sessions", false, "Enable debug mode")
)

func main() {
	// Parse Flags
	flag.Parse()

	// Parse Config
	cfg := mixmaster.ParseConfig(*configFile)

	// Create Device Array
	devices := make(map[string]*mixmaster.MixMasterInstance)

	// Check if show session flag is enabled
	if *showSessions {
		dev, _ := mixmaster.GetDevice()
		// List Devices connected to the computer
		fmt.Println("Device List", dev)
		return
	}

	// Start GUI app and send user a notification that the app has started
	a := app.New()
	a.SendNotification(fyne.NewNotification("Mixmaster", "App has been started"))
	// Create a application window
	w := a.NewWindow("MixMaster")

	// Check if the app is running on a desktop and start system tray
	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("MixMaster",
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}),
			// Scan for devices on demand
			fyne.NewMenuItem("Device Scan", func() {
				// Get a list of all devices pluged into computer
				dev, _ := mixmaster.GetDevice()

				// Loop Through Devices in the config and see if they are connected to the computer
				for deviceName, device := range cfg.Devices {
					tempDevice, err := mixmaster.NewMixMaster(dev, device.SerialNumber)
					if err != nil {
						continue
					}
					devices[deviceName] = tempDevice
				}
			}))

		desk.SetSystemTrayMenu(m)
	} else {
		// End program and alert user that the device only works on desktop
		panic("This application only works on desktop devices")
	}

	// Set window name
	w.SetContent(widget.NewLabel("Mixmaster"))
	// intercept application close and minimize to system tray instead
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	/////////////////////////////// GUI Section   /////////////////////////////////

	// Connect application window to ui library
	devicesPage := ui.NewDeviceListView(w)
	//settingsPage := ui.NewSettingsView()

	// Create ui tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Devices", devicesPage),
		//container.NewTabItem("Settings", settingsPage),
	)
	// connect ui tabs to the application window
	w.SetContent(tabs)

	///////////////////////////////////////////////////////////////////////////////

	go func() {
		// Create Pulse Client
		pulseClient, err := mixmaster.CreatePulseClient("MixMaster")
		if err != nil {
			panic("could not creating Pulse client")
		}

		// Set up mpris
		mprisClient, err := mixmaster.MprisInitialize()
		if err != nil {
			panic("could not creating mpirs client")
		}

		// Get a list of all devices pluged into computer
		dev, _ := mixmaster.GetDevice()

		// Loop Through Devices in the config and see if they are connected to the computer
		for deviceName, device := range cfg.Devices {
			tempDevice, err := mixmaster.NewMixMaster(dev, device.SerialNumber)
			if err != nil {
				continue
			}
			devices[deviceName] = tempDevice
		}

		for {
			var deviceData []*mixmaster.ParsedAudioData

			for deviceName, device := range devices {
				test := cfg.Devices[deviceName]

				data, err := device.Pull(&test)
				if err != nil {
					a.SendNotification(fyne.NewNotification("Mixmaster", "Lost Connection with Device"))
					delete(devices, deviceName)
					continue
				}
				// Add Data to the device data array
				deviceData = append(deviceData, data)
			}

			// Join all the different devices together
			dat := mixmaster.JoinDeviceData(deviceData)

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

			// Update App Volume
			pulseSessions.ChangeAppVolume(dat.PulseApps, pulseClient)
			// Update Master output Volume
			pulseSessions.ChangeMasterVolume(dat.MasterOuputs, pulseClient)
			// Pause, Play, skip song controls
			mpirsSessions.MediaControls(dat.MpirsApps, mprisClient)

			// Slow down device pulling
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Start GUI
	a.Run()
}
