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
type VolumeEntry struct {
	NameEntry   *widget.Entry
	NumberEntry *widget.Entry
}

type AppControlEntry struct {
	NameEntry      *widget.Entry
	BackEntry      *widget.Entry
	PlayPauseEntry *widget.Entry
	NextEntry      *widget.Entry
}

type AppControlNumber struct {
	Back      int
	PlayPause int
	Next      int
}

func DevicePage(w fyne.Window, cfg *Config, configPath *string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance) fyne.CanvasObject {
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

				w.SetContent(EditorPage(w, cfg, configPath, name, deviceList, connectedDevices, devices))
			})

			deviceButtons.Add(btn)
		}

		deviceButtons.Refresh()
	}
	// Initial population
	updateDeviceButtons()

	// Listen for changes in the device list or connection status
	deviceList.AddListener(binding.NewDataListener(func() {
		updateDeviceButtons()
	}))
	connectedDevices.AddListener(binding.NewDataListener(func() {
		updateDeviceButtons()
	}))

	addBtn := widget.NewButton("Add Device", func() {
		fmt.Println("Add device clicked")
	})

	deviceScan := widget.NewButton("Scan for Devices", func() {
		ScanForDevices(cfg, deviceList, connectedDevices, devices)
	})

	return container.NewBorder(nil, addBtn, nil, deviceScan, deviceButtons)
}

func EditorPage(w fyne.Window, cfg *Config, configPath *string, name string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance) fyne.CanvasObject {
	backBtn := widget.NewButton("Back", func() {
		fmt.Println("Back clicked")
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices))
	})

	// --- Get the device from config ---
	device, ok := cfg.Devices[name]
	if !ok {
		return container.NewBorder(backBtn, nil, nil, nil, container.NewCenter(widget.NewLabel("No device selected")))
	}

	deviceName := widget.NewEntry()
	deviceName.SetText(name)
	deviceName.SetPlaceHolder("Device Name")

	appVolumeEntries := &[]VolumeEntry{}
	appVolumeList := container.NewVBox()
	masterVolumeEntries := &[]VolumeEntry{}
	masterVolumeList := container.NewVBox()
	appControlEntries := &[]AppControlEntry{}
	appControlList := container.NewVBox()

	// --- Populate existing apps ---
	if len(device.AppVolumeControls) > 0 {
		for appName, num := range device.AppVolumeControls {
			addAppVolumeEntry(appName, &num, appVolumeEntries, appVolumeList)
		}
	}
	if len(device.MasterVolumeControls) > 0 {
		for appName, num := range device.MasterVolumeControls {
			addAppVolumeEntry(appName, &num, masterVolumeEntries, masterVolumeList)
		}
	}
	if len(device.AppMediaControls) > 0 {
		for appName, num := range device.AppMediaControls {
			fmt.Println(appName, num)
			addAppControlEntry(appName, &num, appControlEntries, appControlList)
		}
	}

	addAppVolumeButton := widget.NewButton("Add App Volume", func() {
		addAppVolumeEntry("", nil, appVolumeEntries, appVolumeList)
	})
	addMasterVolumeButton := widget.NewButton("Add Master Output Volume", func() {
		addAppVolumeEntry("", nil, masterVolumeEntries, masterVolumeList)
	})
	addAppControlButton := widget.NewButton("Add Media Controls", func() {
		addAppControlEntry("", nil, appControlEntries, appControlList)
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
		newAppVolume := make(map[string]int)
		for _, e := range *appVolumeEntries {
			if e.NameEntry.Text != "" {
				if num, err := strconv.Atoi(e.NumberEntry.Text); err == nil && e.NameEntry.Text != "" {
					newAppVolume[e.NameEntry.Text] = num
				} else {
					newAppVolume[e.NameEntry.Text] = -1
				}
			}
		}
		device.AppVolumeControls = newAppVolume

		// rebuild MasterVolumeControls
		newMasterVolume := make(map[string]int)
		for _, e := range *masterVolumeEntries {
			if e.NameEntry.Text != "" {
				if num, err := strconv.Atoi(e.NumberEntry.Text); err == nil && e.NameEntry.Text != "" {
					newMasterVolume[e.NameEntry.Text] = num
				} else {
					newMasterVolume[e.NameEntry.Text] = -1
				}
			}
		}
		device.MasterVolumeControls = newMasterVolume

		// rebuild AppMediaControls
		newControlsApps := make(map[string]mpirsData)
		for _, e := range *appControlEntries {
			var mpirsData mpirsData
			if e.NameEntry.Text != "" {
				if num, err := strconv.Atoi(e.BackEntry.Text); err == nil && e.BackEntry.Text != "" {
					mpirsData.Back = num
				} else {
					mpirsData.Back = -1
				}
				if num, err := strconv.Atoi(e.PlayPauseEntry.Text); err == nil && e.PlayPauseEntry.Text != "" {
					mpirsData.PlayPause = num
				} else {
					mpirsData.PlayPause = -1
				}
				if num, err := strconv.Atoi(e.NextEntry.Text); err == nil && e.NextEntry.Text != "" {
					mpirsData.Next = num
				} else {
					mpirsData.Next = -1
				}

				newControlsApps[e.NameEntry.Text] = mpirsData
			}
		}
		device.AppMediaControls = newControlsApps
		cfg.Devices[name] = device

		cfg.SaveConfig(configPath)
		ScanForDevices(cfg, deviceList, connectedDevices, devices)
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices))
	})

	// --- Center content ---
	centerContent := container.NewVBox(
		widget.NewLabel("Edit Device"),
		deviceName,
		widget.NewSeparator(),
		widget.NewLabel("App Volumes:"),
		appVolumeList,
		addAppVolumeButton,
		widget.NewSeparator(),
		widget.NewLabel("Master Output Volume Controls:"),
		masterVolumeList,
		addMasterVolumeButton,
		widget.NewSeparator(),
		widget.NewLabel("App Media Controls:"),
		appControlList,
		addAppControlButton,
		widget.NewSeparator(),
		saveButton,
	)

	return container.NewBorder(backBtn, nil, nil, nil, centerContent)
}

func refreshAppVolumeList(appVolumeEntries *[]VolumeEntry, appVolumeList *fyne.Container) {
	children := []fyne.CanvasObject{}
	for _, e := range *appVolumeEntries {
		entry := e
		row := container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(200, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.NumberEntry.MinSize().Height)), entry.NumberEntry),
			widget.NewButton("Remove", func() {
				// remove from slice
				newList := []VolumeEntry{}
				for _, a := range *appVolumeEntries {
					if a != entry {
						newList = append(newList, a)
					}
				}
				*appVolumeEntries = newList
				refreshAppVolumeList(appVolumeEntries, appVolumeList)
			}),
		)
		children = append(children, row)
	}
	appVolumeList.Objects = children
	appVolumeList.Refresh()
}

func refreshMasterVolumeList(masterVolumeEntries *[]VolumeEntry, masterVolumeList *fyne.Container) {
	children := []fyne.CanvasObject{}
	for _, e := range *masterVolumeEntries {
		entry := e
		row := container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(200, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.NumberEntry.MinSize().Height)), entry.NumberEntry),
			widget.NewButton("Remove", func() {
				// remove from slice
				newList := []VolumeEntry{}
				for _, a := range *masterVolumeEntries {
					if a != entry {
						newList = append(newList, a)
					}
				}
				*masterVolumeEntries = newList
				refreshAppVolumeList(masterVolumeEntries, masterVolumeList)
			}),
		)
		children = append(children, row)
	}
	masterVolumeList.Objects = children
	masterVolumeList.Refresh()
}

func refreshAppControlList(appControlEntries *[]AppControlEntry, appControlList *fyne.Container) {
	children := []fyne.CanvasObject{}
	for _, e := range *appControlEntries {
		entry := e
		row := container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(200, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.BackEntry.MinSize().Height)), entry.BackEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.PlayPauseEntry.MinSize().Height)), entry.PlayPauseEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.NextEntry.MinSize().Height)), entry.NextEntry),
			widget.NewButton("Remove", func() {
				// remove from slice
				newList := []AppControlEntry{}
				for _, a := range *appControlEntries {
					if a != entry {
						newList = append(newList, a)
					}
				}
				*appControlEntries = newList
				refreshAppControlList(appControlEntries, appControlList)
			}),
		)
		children = append(children, row)
	}
	appControlList.Objects = children
	appControlList.Refresh()
}

func addAppVolumeEntry(appName string, appNumber *int, appEntries *[]VolumeEntry, appList *fyne.Container) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("App Name")
	nameEntry.SetText(appName)

	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("App Number")
	if appNumber != nil {
		if *appNumber != -1 {
			numberEntry.SetText(fmt.Sprintf("%d", *appNumber))
		}
	}

	*appEntries = append(*appEntries, VolumeEntry{nameEntry, numberEntry})
	refreshAppVolumeList(appEntries, appList)
}

func addAppControlEntry(appName string, controlNumbers *mpirsData, appControlEntries *[]AppControlEntry, appControlList *fyne.Container) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("App Name")
	nameEntry.SetText(appName)

	// Back
	backEntry := widget.NewEntry()
	backEntry.SetPlaceHolder("Back Number")
	// PlayPause
	playpauseEntry := widget.NewEntry()
	playpauseEntry.SetPlaceHolder("Play/Pause Number")
	// Next
	nextEntry := widget.NewEntry()
	nextEntry.SetPlaceHolder("Next Number")

	if controlNumbers != nil {
		if controlNumbers.Back != -1 {
			backEntry.SetText(fmt.Sprintf("%d", controlNumbers.Back))
		}
		if controlNumbers.PlayPause != -1 {
			playpauseEntry.SetText(fmt.Sprintf("%d", controlNumbers.PlayPause))
		}
		if controlNumbers.Next != -1 {
			nextEntry.SetText(fmt.Sprintf("%d", controlNumbers.Next))
		}
	}

	*appControlEntries = append(*appControlEntries, AppControlEntry{nameEntry, backEntry, playpauseEntry, nextEntry})
	refreshAppControlList(appControlEntries, appControlList)
}

func addMasterVolumeEntry(masterName string, masterNumber *int, masterEntries *[]VolumeEntry, appList *fyne.Container) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Output Name")
	nameEntry.SetText(masterName)

	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("Slider Number")
	if masterNumber != nil {
		if *masterNumber != -1 {
			numberEntry.SetText(fmt.Sprintf("%d", *masterNumber))
		}
	}

	*masterEntries = append(*masterEntries, VolumeEntry{nameEntry, numberEntry})
	refreshAppVolumeList(masterEntries, appList)
}
