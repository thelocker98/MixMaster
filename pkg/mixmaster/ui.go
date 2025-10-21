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

type ControlsCount struct {
	numOfSlidders int
	numOfButtons  int
}

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

func DevicePage(w fyne.Window, cfg *Config, configPath *string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, serialNumbers *map[string]string, pulseSessions *PulseSessions, mpirsSessions *MpirsSessions) fyne.CanvasObject {
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
				w.SetContent(EditorPage(w, cfg, configPath, name, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
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
		w.SetContent(EditorPage(w, cfg, configPath, "", deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
	})

	scanBtn := widget.NewButtonWithIcon("Scan Devices", theme.ViewRefreshIcon(), func() {
		ScanForDevices(cfg, deviceList, connectedDevices, devices, serialNumbers)
	})

	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		w.SetContent(SettingsPage(w, cfg, configPath, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
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

func EditorPage(w fyne.Window, cfg *Config, configPath *string, name string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, serialNumbers *map[string]string, pulseSessions *PulseSessions, mpirsSessions *MpirsSessions) fyne.CanvasObject {
	title := canvas.NewText("Device Editor", color.White) //theme.ForegroundColor())
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
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
	deviceSerial.SetPlaceHolder("Serial Number")

	appVolumeEntries := &[]VolumeEntry{}
	appVolumeList := container.NewVBox()
	masterVolumeEntries := &[]VolumeEntry{}
	masterVolumeList := container.NewVBox()
	appControlEntries := &[]AppControlEntry{}
	appControlList := container.NewVBox()

	// --- Populate existing apps ---
	if len(device.AppVolumeControls) > 0 {
		for appName, num := range device.AppVolumeControls {
			addAppVolumeEntry(w, appName, &num, appVolumeEntries, appVolumeList, pulseSessions)
		}
	}
	if len(device.MasterVolumeControls) > 0 {
		for appName, num := range device.MasterVolumeControls {
			addMasterVolumeEntry(w, appName, &num, masterVolumeEntries, masterVolumeList, pulseSessions)
		}
	}
	if len(device.AppMediaControls) > 0 {
		for appName, num := range device.AppMediaControls {
			addAppControlEntry(w, appName, &num, appControlEntries, appControlList, mpirsSessions)
		}
	}

	addAppVolumeButton := widget.NewButton("Add App Volume", func() {
		addAppVolumeEntry(w, "", nil, appVolumeEntries, appVolumeList, pulseSessions)
	})
	addMasterVolumeButton := widget.NewButton("Add Master Output Volume", func() {
		addMasterVolumeEntry(w, "", nil, masterVolumeEntries, masterVolumeList, pulseSessions)
	})
	addAppControlButton := widget.NewButton("Add Media Controls", func() {
		addAppControlEntry(w, "", nil, appControlEntries, appControlList, mpirsSessions)
	})

	saveButton := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		if deviceName.Text == "" || deviceSerial.Text == "" {
			fmt.Println("Invalid name or serial number")
			return
		}

		newName := deviceName.Text

		// update device name if changed
		if newName != name {
			cfg.Devices[newName] = cfg.Devices[name]
			delete(cfg.Devices, name)
			name = newName
		}

		// Save Serial Numbers
		device.SerialNumber = deviceSerial.Text

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
		ScanForDevices(cfg, deviceList, connectedDevices, devices, serialNumbers)
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
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
					ScanForDevices(cfg, deviceList, connectedDevices, devices, serialNumbers)
					cfg.SaveConfig(configPath)
					w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
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

	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		// declare globally so you can close it inside the list
		var d dialog.Dialog

		// Create a list of serial numbers for selection
		var serialOptions []string
		for serial := range *serialNumbers {
			serialOptions = append(serialOptions, serial)
		}

		// List widget for dialog
		serialList := widget.NewList(
			func() int { return len(serialOptions) },
			func() fyne.CanvasObject {
				return widget.NewButton("template", nil)
			},
			func(i widget.ListItemID, o fyne.CanvasObject) {
				btn := o.(*widget.Button)
				serial := serialOptions[i]
				// Display Serial Number and Name Entry
				btn.SetText(fmt.Sprintf("%s - %s", serial, (*serialNumbers)[serial]))
				btn.OnTapped = func() {
					// Set Serial Number Value to chosen value
					deviceSerial.SetText(serial)
					// Hide dialog
					d.Hide()
				}
			},
		)

		// Create and show the dialog
		d = dialog.NewCustom("Avalible Devices", "Close", serialList, w)
		d.Resize(fyne.NewSize(400, 300))
		d.Show()
	})

	deviceConfigCard := container.NewVBox(
		container.NewHBox(createSectionHeader("Device Configuration"), widget.NewButtonWithIcon("", theme.QuestionIcon(), func() { fmt.Println("help") })),
		widget.NewSeparator(),
		container.NewVBox(
			container.NewBorder(nil, nil, widget.NewLabel("Device Name:"), nil, deviceName),
			container.NewBorder(nil, nil, widget.NewLabel("Serial Number:"), searchButton, deviceSerial),
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

func SettingsPage(w fyne.Window, cfg *Config, configPath *string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, serialNumbers *map[string]string, pulseSessions *PulseSessions, mpirsSessions *MpirsSessions) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		cfg.SaveConfig(configPath)
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
	})

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		w.SetContent(DevicePage(w, cfg, configPath, deviceList, connectedDevices, devices, serialNumbers, pulseSessions, mpirsSessions))
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

func refreshAppVolumeList(w fyne.Window, appVolumeEntries *[]VolumeEntry, appVolumeList *fyne.Container, pulseSessions *PulseSessions) {
	children := []fyne.CanvasObject{}
	appNames := []string{}
	for name, _ := range pulseSessions.Apps {
		appNames = append(appNames, name)
	}
	appNames = append(appNames, "unmapped")

	for _, e := range *appVolumeEntries {
		entry := e

		row := container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			searchButtonPopup(w, "Current Apps Avalible", appNames, entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(280, entry.NumberEntry.MinSize().Height)), entry.NumberEntry),
			widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
				// remove from slice
				newList := []VolumeEntry{}
				for _, a := range *appVolumeEntries {
					if a != entry {
						newList = append(newList, a)
					}
				}
				*appVolumeEntries = newList
				refreshAppVolumeList(w, appVolumeEntries, appVolumeList, pulseSessions)
			}),
		)
		children = append(children, row)
	}
	appVolumeList.Objects = children
	appVolumeList.Refresh()
}

func refreshMasterVolumeList(w fyne.Window, masterVolumeEntries *[]VolumeEntry, masterVolumeList *fyne.Container, pulseSessions *PulseSessions) {
	children := []fyne.CanvasObject{}

	outputNames := []string{}
	for name, _ := range pulseSessions.Masters {
		outputNames = append(outputNames, name)
	}
	outputNames = append(outputNames, "master")

	for _, e := range *masterVolumeEntries {
		entry := e
		row := container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			searchButtonPopup(w, "Current Outputs Avalible", outputNames, entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(280, entry.NumberEntry.MinSize().Height)), entry.NumberEntry),
			widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
				// remove from slice
				newList := []VolumeEntry{}
				for _, a := range *masterVolumeEntries {
					if a != entry {
						newList = append(newList, a)
					}
				}
				*masterVolumeEntries = newList
				refreshMasterVolumeList(w, masterVolumeEntries, masterVolumeList, pulseSessions)
			}),
		)
		children = append(children, row)
	}
	masterVolumeList.Objects = children
	masterVolumeList.Refresh()
}

func refreshAppControlList(w fyne.Window, appControlEntries *[]AppControlEntry, appControlList *fyne.Container, mpirsSessions *MpirsSessions) {
	children := []fyne.CanvasObject{}
	mpirsAppNames := []string{}
	for name, _ := range *mpirsSessions {
		mpirsAppNames = append(mpirsAppNames, name)
	}

	for _, e := range *appControlEntries {
		entry := e
		row := container.NewHBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(150, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			searchButtonPopup(w, "Current Apps Avalible", mpirsAppNames, entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(90, entry.BackEntry.MinSize().Height)), entry.BackEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(90, entry.PlayPauseEntry.MinSize().Height)), entry.PlayPauseEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(90, entry.NextEntry.MinSize().Height)), entry.NextEntry),
			widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
				// remove from slice
				newList := []AppControlEntry{}
				for _, a := range *appControlEntries {
					if a != entry {
						newList = append(newList, a)
					}
				}
				*appControlEntries = newList
				refreshAppControlList(w, appControlEntries, appControlList, mpirsSessions)
			}),
		)
		children = append(children, row)
	}
	appControlList.Objects = children
	appControlList.Refresh()
}

func addAppVolumeEntry(w fyne.Window, appName string, appNumber *int, appEntries *[]VolumeEntry, appList *fyne.Container, pulseSessions *PulseSessions) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("App Name")
	nameEntry.SetText(appName)

	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("Slidder")
	if appNumber != nil {
		if *appNumber != -1 {
			numberEntry.SetText(fmt.Sprintf("%d", *appNumber))
		}
	}

	levelWidget := widget.NewProgressBar()
	levelWidget.SetValue(0.5)

	*appEntries = append(*appEntries, VolumeEntry{nameEntry, numberEntry})
	refreshAppVolumeList(w, appEntries, appList, pulseSessions)
}

func addMasterVolumeEntry(w fyne.Window, masterName string, masterNumber *int, masterEntries *[]VolumeEntry, appList *fyne.Container, pulseSessions *PulseSessions) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Output Name")
	nameEntry.SetText(masterName)

	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("Slidder")
	if masterNumber != nil {
		if *masterNumber != -1 {
			numberEntry.SetText(fmt.Sprintf("%d", *masterNumber))
		}
	}

	levelWidget := widget.NewProgressBar()
	levelWidget.SetValue(0.5)

	*masterEntries = append(*masterEntries, VolumeEntry{nameEntry, numberEntry})
	refreshMasterVolumeList(w, masterEntries, appList, pulseSessions)
}

func addAppControlEntry(w fyne.Window, appName string, controlNumbers *mpirsData, appControlEntries *[]AppControlEntry, appControlList *fyne.Container, mpirsSessions *MpirsSessions) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("App Name")
	nameEntry.SetText(appName)

	// Back
	backEntry := widget.NewEntry()
	backEntry.SetPlaceHolder("Back")
	// PlayPause
	playpauseEntry := widget.NewEntry()
	playpauseEntry.SetPlaceHolder("Play/Pause")
	// Next
	nextEntry := widget.NewEntry()
	nextEntry.SetPlaceHolder("Next")

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
	refreshAppControlList(w, appControlEntries, appControlList, mpirsSessions)
}

func searchButtonPopup(w fyne.Window, title string, textInput []string, elementToChange *widget.Entry) *widget.Button {
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		// declare globally so you can close it inside the list
		var d dialog.Dialog

		// List widget for dialog
		serialList := widget.NewList(
			func() int { return len(textInput) },
			func() fyne.CanvasObject {
				return widget.NewButton("template", nil)
			},
			func(i widget.ListItemID, o fyne.CanvasObject) {
				btn := o.(*widget.Button)
				serial := textInput[i]
				// Display Serial Number and Name Entry
				btn.SetText(fmt.Sprint(serial))
				btn.OnTapped = func() {
					// Set Serial Number Value to chosen value
					elementToChange.SetText(serial)
					// Hide dialog
					d.Hide()
				}
			},
		)

		// Create and show the dialog
		d = dialog.NewCustom(title, "Close", serialList, w)
		d.Resize(fyne.NewSize(400, 300))
		d.Show()
	})

	return searchButton
}

func processSerialNumber(serialNumber string) (ControlsCount, error) {
	var controlsCount ControlsCount

	num, err := strconv.Atoi(serialNumber[3:5])
	if err != nil {
		return ControlsCount{}, err
	}
	controlsCount.numOfSlidders = num

	num, err = strconv.Atoi(serialNumber[6:8])
	if err != nil {
		return ControlsCount{}, err
	}
	controlsCount.numOfButtons = num

	return controlsCount, nil
}
