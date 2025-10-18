package mixmaster

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- App list setup ---
type VolumeEntry struct {
	NameEntry   *widget.Entry
	NumberEntry *widget.Entry
	LevelWidget *widget.ProgressBar
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

func DevicePage(w fyne.Window, cfg *Config, configPath *string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, data *ParsedAudioData) fyne.CanvasObject {
	title := canvas.NewText("MixMaster", color.White)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	deviceButtons := container.NewVBox()

	updateDeviceButtons := func() {
		deviceButtons.Objects = nil // clear old buttons

		names, _ := deviceList.Get()
		for i, name := range names {
			status := "Disconnected"
			color := theme.ColorNameError

			if val, _ := connectedDevices.GetValue(i); val {
				status = "Connected"
				color = theme.ColorNameSuccess
			}

			statusLabel := canvas.NewText(status, theme.Color(color))
			statusLabel.Alignment = fyne.TextAlignTrailing

			btn := widget.NewButton(name, func() {
				w.SetContent(EditorPage(w, cfg, configPath, name, deviceList, connectedDevices, devices, data))
			})

			// Each device row as a horizontal box
			row := container.NewHBox(layout.NewSpacer(), btn, statusLabel, layout.NewSpacer())

			fixedblock := container.NewCenter(row)
			fixedblock.Resize(fyne.NewSize(100, 200))

			deviceButtons.Add(fixedblock)
		}

		deviceButtons.Refresh()
	}

	updateDeviceButtons()

	deviceList.AddListener(binding.NewDataListener(func() { updateDeviceButtons() }))
	connectedDevices.AddListener(binding.NewDataListener(func() { updateDeviceButtons() }))

	addBtn := widget.NewButtonWithIcon("Add Device", theme.ContentAddIcon(), func() {
		w.SetContent(EditorPage(w, cfg, configPath, "", deviceList, connectedDevices, devices, data))
	})

	scanBtn := widget.NewButtonWithIcon("Scan Devices", theme.ViewRefreshIcon(), func() {
		ScanForDevices(cfg, deviceList, connectedDevices, devices)
	})

	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		w.SetContent(SettingsPage(w, cfg, configPath, deviceList, connectedDevices, devices, data))
	})

	deviceBoxTitle := canvas.NewText("Devices", color.White)
	deviceBoxTitle.TextSize = 24
	deviceBoxTitle.TextStyle = fyne.TextStyle{Bold: true}
	deviceBoxTitle.Alignment = fyne.TextAlignCenter

	scrollableList := container.NewVScroll(container.NewVBox(layout.NewSpacer(), deviceBoxTitle, deviceButtons, layout.NewSpacer()))
	scrollableList.SetMinSize(fyne.NewSize(300, 300))

	content := container.NewBorder(
		container.NewHBox(settingsBtn, layout.NewSpacer(), title, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), addBtn, scanBtn),
		nil,
		nil,
		scrollableList,
	)

	return content
}

func EditorPage(w fyne.Window, cfg *Config, configPath *string, name string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, data *ParsedAudioData) fyne.CanvasObject {
	fmt.Println(data)

	title := canvas.NewText("Device Editor", color.White) //theme.ForegroundColor())
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, data))
	})

	// --- Get the device from config ---
	device, ok := cfg.Devices[name]
	if !ok {
		device = DeviceConfig{}
	}

	deviceName := widget.NewEntry()
	deviceName.SetText(name)
	deviceName.SetPlaceHolder("Device Name")

	deviceSerial := widget.NewEntry()
	deviceSerial.SetText(string(device.SerialNumber))
	deviceSerial.SetPlaceHolder("Device Serial Number")

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
			addMasterVolumeEntry(appName, &num, masterVolumeEntries, masterVolumeList)
		}
	}
	if len(device.AppMediaControls) > 0 {
		for appName, num := range device.AppMediaControls {
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

	saveButton := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
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
		device.SerialNumber = deviceSerial.Text
		cfg.Devices[name] = device

		cfg.SaveConfig(configPath)
		ScanForDevices(cfg, deviceList, connectedDevices, devices)
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, data))
	})

	removeDeviceButton := widget.NewButtonWithIcon("Delete Device", theme.DeleteIcon(), func() {
		// Create a custom confirmation dialog
		content := widget.NewLabel("Are you sure you want to delete this device?")

		d := dialog.NewCustomConfirm(
			"Confirm Deletion",
			"Cancel",
			"Delete",
			content,
			func(confirmed bool) {
				if !confirmed {
					delete(cfg.Devices, name)
					ScanForDevices(cfg, deviceList, connectedDevices, devices)
					cfg.SaveConfig(configPath)
					w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, data))
				}
			},
			w,
		)
		d.Show()
	})

	// Helper function to create section headers
	createSectionHeader := func(text string) *canvas.Text {
		header := canvas.NewText(text, color.White)
		header.TextSize = 18
		header.TextStyle = fyne.TextStyle{Bold: true}
		return header
	}

	deviceConfigCard := container.NewVBox(
		createSectionHeader("Device Configuration"),
		widget.NewSeparator(),
		container.NewVBox(
			container.NewBorder(nil, nil, widget.NewLabel("Device Name:"), nil, deviceName),
			container.NewBorder(nil, nil, widget.NewLabel("Serial Number:"), widget.NewButtonWithIcon("", theme.SearchIcon(), func() { fmt.Println("Searching Serial Numbers") }), deviceSerial),
		),
	)

	appVolumeCard := container.NewVBox(
		createSectionHeader("App Volume Configuration"),
		widget.NewSeparator(),
		appVolumeList,
		addAppVolumeButton,
	)

	masterVolumeCard := container.NewVBox(
		createSectionHeader("Master Volume Configuration"),
		widget.NewSeparator(),
		masterVolumeList,
		addMasterVolumeButton,
	)

	mediaControlCard := container.NewVBox(
		createSectionHeader("Media Configuration"),
		widget.NewSeparator(),
		appControlList,
		addAppControlButton,
	)

	// Main content container with padding
	contentContainer := container.NewVBox(
		deviceConfigCard,
		widget.NewLabel(""), // Spacer
		appVolumeCard,
		widget.NewLabel(""), // Spacer
		masterVolumeCard,
		widget.NewLabel(""), // Spacer
		mediaControlCard,
	)

	// Wrap content in a max-width container
	maxWidthContent := container.NewPadded(
		container.NewCenter(
			container.NewStack(
				// This creates a container that won't exceed the specified size
				&fyne.Container{
					Layout: layout.NewStackLayout(),
					Objects: []fyne.CanvasObject{
						widget.NewLabel(""), // Invisible spacer for max width
					},
				},
				contentContainer,
			),
		),
	)

	// Set a reasonable max size for the invisible spacer
	if len(maxWidthContent.Objects) > 0 {
		if centerContainer, ok := maxWidthContent.Objects[0].(*fyne.Container); ok {
			if maxContainer, ok := centerContainer.Objects[0].(*fyne.Container); ok {
				if len(maxContainer.Objects) > 0 {
					maxContainer.Objects[0].(*fyne.Container).Objects[0].Resize(fyne.NewSize(600, 1))
				}
			}
		}
	}

	scrollableList := container.NewVScroll(maxWidthContent)
	//scrollableList.SetMinSize(fyne.NewSize(400, 400))

	return container.NewBorder(
		container.NewHBox(backBtn, layout.NewSpacer(), title, layout.NewSpacer(), removeDeviceButton),
		container.NewHBox(layout.NewSpacer(), saveButton),
		nil,
		nil,
		scrollableList,
	)
}

func SettingsPage(w fyne.Window, cfg *Config, configPath *string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, data *ParsedAudioData) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		cfg.SaveConfig(configPath)
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, data))
	})

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, data))
	})

	// Theme toggle
	themeToggle := widget.NewCheck("Dark Mode", func(value bool) {
		if value {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
		} else {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
		}
	})

	// Create card for settings
	settingsCard := container.NewVBox(
		widget.NewLabel("Appearance"),
		widget.NewSeparator(),
		themeToggle,
	)

	// Wrap in padded, centered, max-width container
	maxWidthContent := container.NewPadded(
		container.NewCenter(
			container.NewMax(
				&fyne.Container{
					Layout: layout.NewMaxLayout(),
					Objects: []fyne.CanvasObject{
						widget.NewLabel(""),
					},
				},
				settingsCard,
			),
		),
	)

	// Set max width
	if len(maxWidthContent.Objects) > 0 {
		if centerContainer, ok := maxWidthContent.Objects[0].(*fyne.Container); ok {
			if maxContainer, ok := centerContainer.Objects[0].(*fyne.Container); ok {
				if len(maxContainer.Objects) > 0 {
					maxContainer.Objects[0].(*fyne.Container).Objects[0].Resize(fyne.NewSize(600, 1))
				}
			}
		}
	}

	settingscontent := container.NewVScroll(maxWidthContent)

	return container.NewBorder(
		container.NewHBox(backBtn, layout.NewSpacer(), title, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), saveBtn),
		nil,
		nil,
		settingscontent,
	)
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
		children = append(children, container.NewVBox(row, entry.LevelWidget, widget.NewSeparator()))
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
				refreshMasterVolumeList(masterVolumeEntries, masterVolumeList)
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

	levelWidget := widget.NewProgressBar()
	levelWidget.SetValue(0.5)

	*appEntries = append(*appEntries, VolumeEntry{nameEntry, numberEntry, levelWidget})
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

	levelWidget := widget.NewProgressBar()
	levelWidget.SetValue(0.5)

	*masterEntries = append(*masterEntries, VolumeEntry{nameEntry, numberEntry, levelWidget})
	refreshMasterVolumeList(masterEntries, appList)
}
