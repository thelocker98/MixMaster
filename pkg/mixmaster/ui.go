package mixmaster

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// --- App list setup ---
type AppEntry struct {
	NameEntry   *widget.Entry
	NumberEntry *widget.Entry
}

func DevicePage(w fyne.Window, cfg *Config, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance) fyne.CanvasObject {
	// Create a container to hold buttons for each device
	deviceButtons := container.NewVBox()

	updateDeviceButtons := func() {
		deviceButtons.Objects = nil // clear old buttons

		names, _ := deviceList.Get()
		for i, name := range names {
			status := "disconnected"
			if val, _ := connectedDevices.GetValue(i); val {
				status = "connected"
			}

			// Create a button for each device
			btn := widget.NewButton(fmt.Sprintf("%s - %s", name, status), func() {
				fmt.Println("Clicked:", name)

				w.SetContent(EditorPage(w, cfg, name, deviceList, connectedDevices, devices))
			})

			deviceButtons.Add(btn)
		}

		deviceButtons.Refresh()
	}

	// Listen for changes in the device list or connection status
	deviceList.AddListener(binding.NewDataListener(func() {
		updateDeviceButtons()
	}))
	connectedDevices.AddListener(binding.NewDataListener(func() {
		updateDeviceButtons()
	}))

	// Initial population
	updateDeviceButtons()

	addBtn := widget.NewButton("Add Device", func() {
		fmt.Println("Add device clicked")
	})

	deviceScan := widget.NewButton("Scan for Devices", func() {
		ScanForDevices(cfg, deviceList, connectedDevices, devices)
	})

	return container.NewBorder(nil, addBtn, nil, deviceScan, deviceButtons)
}

func EditorPage(w fyne.Window, cfg *Config, name string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance) fyne.CanvasObject {
	addBtn := widget.NewButton("Back", func() {
		fmt.Println("Back clicked")
		w.SetContent(DevicePage(w, cfg, deviceList, connectedDevices, devices))
	})

	// --- Get the device from config ---
	device, ok := cfg.Devices[name]
	if !ok {
		return container.NewBorder(addBtn, nil, nil, nil, container.NewCenter(widget.NewLabel("No device selected")))
	}

	deviceName := widget.NewEntry()
	deviceName.SetText(name)
	deviceName.SetPlaceHolder("Device Name")

	appEntries := []*AppEntry{}
	appList := container.NewVBox()

	refreshAppList := func() {
		children := []fyne.CanvasObject{}
		for _, e := range appEntries {
			entry := e
			row := container.NewHBox(
				container.New(layout.NewGridWrapLayout(fyne.NewSize(200, entry.NameEntry.MinSize().Height)), entry.NameEntry),
				container.New(layout.NewGridWrapLayout(fyne.NewSize(100, entry.NumberEntry.MinSize().Height)), entry.NumberEntry),
				widget.NewButton("Remove", func() {
					// remove from slice
					newList := []*AppEntry{}
					for _, a := range appEntries {
						if a != entry {
							newList = append(newList, a)
						}
					}
					appEntries = newList
				}),
			)
			children = append(children, row)
		}
		appList.Objects = children
		appList.Refresh()
	}

	addAppRow := func(appName string, appNumber int) {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("App Name")
		nameEntry.SetText(appName)

		numberEntry := widget.NewEntry()
		numberEntry.SetPlaceHolder("App Number")
		numberEntry.SetText(fmt.Sprintf("%d", appNumber))

		appEntries = append(appEntries, &AppEntry{nameEntry, numberEntry})
		refreshAppList()
	}

	// --- Populate existing apps ---
	if len(device.AppVolumeControls) > 0 {
		for appName, num := range device.AppVolumeControls {
			addAppRow(appName, num)
		}
	} else {
		addAppRow("", 0)
	}

	addAppButton := widget.NewButton("Add App", func() {
		addAppRow("", 0)
	})

	saveButton := widget.NewButton("Save", func() {
		newName := deviceName.Text

		// update device name if changed
		if newName != name {
			cfg.Devices[newName] = cfg.Devices[name]
			delete(cfg.Devices, name)
			name = newName
		}

		// rebuild AppVolumeControls
		newApps := make(map[string]int)
		for _, e := range appEntries {
			if e.NameEntry.Text != "" {
				if num, err := strconv.Atoi(e.NumberEntry.Text); err == nil {
					newApps[e.NameEntry.Text] = num
				}
			}
		}
		device.AppVolumeControls = newApps
		cfg.Devices[name] = device

		fmt.Println("Saved device:", name)
		w.SetContent(DevicePage(w, cfg, deviceList, connectedDevices, devices))
	})

	// --- Center content ---
	centerContent := container.NewVBox(
		widget.NewLabel("Edit Device"),
		deviceName,
		widget.NewSeparator(),
		widget.NewLabel("Apps:"),
		appList,
		addAppButton,
		widget.NewSeparator(),
		saveButton,
	)

	return container.NewBorder(addBtn, nil, nil, nil, centerContent)
}
