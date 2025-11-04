package mixmaster

import (
	"fmt"
	"image/color"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
)

var serialNumberString string
var SerialNumbers = make(map[string]string)

type ControlsCount struct {
	numOfSlidders int
	numOfButtons  int
}

// --- App list setup ---
type VolumeEntry struct {
	NameEntry  *widget.Entry
	NumberList *widget.Select
}
type AppControlEntry struct {
	NameEntry     *widget.Entry
	BackList      *widget.Select
	PlayPauseList *widget.Select
	NextList      *widget.Select
}

type AppControlNumber struct {
	Back      int
	PlayPause int
	Next      int
}

var cfgGlobal *Config
var configPathGlobal *string


func InitializeUI(cfg *Config, configPath *string) {
	cfgGlobal = cfg
	configPathGlobal = configPath
	
}

func DevicePage(a fyne.App, w fyne.Window, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, pulseSessions *PulseSessions, mpirsSessions *MpirsSessions) fyne.CanvasObject {
	title := canvas.NewText("MixMaster", color.White)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	deviceButtons := container.NewVBox()

	updateDeviceButtons := func() {
		deviceButtons.Objects = nil // clear old buttons

		names, _ := deviceList.Get()
		if len(names) == 0 {
			// Each device row as a horizontal box
			row := container.NewHBox(layout.NewSpacer(), widget.NewTextGridFromString("No Devices"), layout.NewSpacer())

			fixedblock := container.NewCenter(row)
			fixedblock.Resize(fyne.NewSize(100, 200))

			deviceButtons.Add(container.NewHBox(layout.NewSpacer(), widget.NewButtonWithIcon("Add Device", theme.ContentAddIcon(), func() {
				w.SetContent(EditorPage(a, w, "", deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
			}), layout.NewSpacer()))
		}

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
				w.SetContent(EditorPage(a, w, name, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
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
		w.SetContent(EditorPage(a, w, "", deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
	})

	scanBtn := widget.NewButtonWithIcon("Scan Devices", theme.ViewRefreshIcon(), func() {
		ScanForDevices(a, cfgGlobal, deviceList, connectedDevices, devices, &SerialNumbers)
	})

	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		w.SetContent(SettingsPage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
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

func EditorPage(a fyne.App, w fyne.Window, name string, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, pulseSessions *PulseSessions, mpirsSessions *MpirsSessions) fyne.CanvasObject {
	title := canvas.NewText("Device Editor", color.White) //theme.ForegroundColor())
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		w.SetContent(DevicePage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
	})

	// --- Get the device from config ---
	device, ok := cfgGlobal.Devices[name]
	if !ok {
		device = DeviceConfig{}
	}

	deviceName := widget.NewEntry()
	deviceName.SetText(name)
	deviceName.SetPlaceHolder("Device Name")

	deviceSerial := widget.NewEntry()
	deviceSerial.SetText(string(device.SerialNumber))
	deviceSerial.SetPlaceHolder("Serial Number")
	serialNumberString = deviceSerial.Text

	appVolumeEntries := &[]VolumeEntry{}
	appVolumeList := container.NewVBox()
	masterVolumeEntries := &[]VolumeEntry{}
	masterVolumeList := container.NewVBox()
	appControlEntries := &[]AppControlEntry{}
	appControlList := container.NewVBox()

	// --- Populate existing apps ---
	if len(device.AppVolumeControls) > 0 {
		for appName, num := range device.AppVolumeControls {
			addAppVolumeEntry(w, appName, num, appVolumeEntries, appVolumeList, pulseSessions)
		}
	}
	if len(device.MasterVolumeControls) > 0 {
		for appName, num := range device.MasterVolumeControls {
			addMasterVolumeEntry(w, appName, num, masterVolumeEntries, masterVolumeList, pulseSessions)
		}
	}
	if len(device.AppMediaControls) > 0 {
		for appName, num := range device.AppMediaControls {
			addAppControlEntry(w, appName, &num, appControlEntries, appControlList, mpirsSessions)
		}
	}

	addAppVolumeButton := widget.NewButton("Add App Volume", func() {
		addAppVolumeEntry(w, "", -1, appVolumeEntries, appVolumeList, pulseSessions)
	})
	addMasterVolumeButton := widget.NewButton("Add Master Output Volume", func() {
		addMasterVolumeEntry(w, "", -1, masterVolumeEntries, masterVolumeList, pulseSessions)
	})
	addAppControlButton := widget.NewButton("Add Media Controls", func() {
		addAppControlEntry(w, "", nil, appControlEntries, appControlList, mpirsSessions)
	})

	exportButton := widget.NewButtonWithIcon("Export Device", theme.DownloadIcon(), func() {
		deviceData := cfgGlobal.Devices[name]

		exportData, _ := yaml.Marshal(deviceData)
		// Create file save dialog
		saveDialog := dialog.NewFileSave(
			func(uc fyne.URIWriteCloser, err error) {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				if uc == nil {
					// user cancelled
					return
				}
				defer uc.Close()

				// Write data to chosen file
				_, err = uc.Write([]byte(exportData))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				// Show confirmation
				dialog.ShowInformation("Export Complete", "File saved successfully!", w)
			}, w,
		)

		// Export name
		saveDialog.SetFileName(name + ".yaml")

		// Optionally, restrict file types (e.g., only text files)
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".yaml"}))

		saveDialog.Show()
	})

	importButton := widget.NewButtonWithIcon("Import Device", theme.UploadIcon(), func() {
		// Create file open dialog
		openDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uc == nil {
				// user cancelled
				return
			}
			defer uc.Close()

			// Extract file name (without extension) to use as device name
			fileURI := uc.URI()
			fileName := filepath.Base(fileURI.Path())
			deviceName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			// Read YAML data
			data, err := io.ReadAll(uc)
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to read file: %w", err), w)
				return
			}

			// Unmarshal YAML into a device struct
			var device DeviceConfig
			err = yaml.Unmarshal(data, &device)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid YAML format: %w", err), w)
				return
			}

			// Append or overwrite in config
			cfgGlobal.Devices[deviceName] = device

			// Show confirmation
			dialog.ShowInformation("Import Complete", fmt.Sprintf("Device '%s' imported successfully!", deviceName), w)

			// save and refresh home screen
			cfgGlobal.SaveConfig(configPathGlobal)
			ScanForDevices(a, cfgGlobal, deviceList, connectedDevices, devices, &SerialNumbers)
			w.SetContent(DevicePage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
		}, w)

		// Restrict file types
		openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".yaml"}))
		openDialog.Show()
	})

	saveButton := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		_, err := processSerialNumber(deviceSerial.Text)
		if deviceName.Text == "" || deviceSerial.Text == "" || err != nil {
			content := widget.NewLabel("Error Parsing Device Name or Device Serial Name")
			dialog.NewCustom("Parsing Error", "Ok", content, w).Show()
			return
		}

		newName := deviceName.Text

		// update device name if changed
		if newName != name {
			cfgGlobal.Devices[newName] = cfgGlobal.Devices[name]
			delete(cfgGlobal.Devices, name)
			delete(*devices, name)
			name = newName
		}

		// Save Serial Numbers
		device.SerialNumber = deviceSerial.Text

		// rebuild AppVolumeControls
		newAppVolume := make(map[string]int)
		for _, e := range *appVolumeEntries {
			if e.NameEntry.Text != "" || len(e.NumberList.Selected) < 10 {
				if e.NumberList.Selected == "none" {
					newAppVolume[e.NameEntry.Text] = -1
				} else {
					num, _ := strconv.Atoi(e.NumberList.Selected[9:])
					newAppVolume[e.NameEntry.Text] = num - 1 // minus one to bring it back to 0 as the first
				}
			}
		}
		device.AppVolumeControls = newAppVolume

		// rebuild MasterVolumeControls
		newMasterVolume := make(map[string]int)
		for _, e := range *masterVolumeEntries {
			if e.NameEntry.Text != "" {
				if e.NumberList.Selected == "none" || len(e.NumberList.Selected) < 10 {
					newMasterVolume[e.NameEntry.Text] = -1
				} else {
					num, _ := strconv.Atoi(e.NumberList.Selected[9:])
					newMasterVolume[e.NameEntry.Text] = num - 1 // minus one to bring it back to 0 as the first
				}
			}
		}
		device.MasterVolumeControls = newMasterVolume

		// rebuild AppMediaControls
		newControlsApps := make(map[string]mpirsData)
		for _, e := range *appControlEntries {
			var mpirsData mpirsData
			if e.NameEntry.Text != "" {
				if e.BackList.Selected == "none" || len(e.BackList.Selected) < 9 {
					mpirsData.Back = -1
				} else {
					num, _ := strconv.Atoi(e.BackList.Selected[8:])
					mpirsData.Back = num - 1 // minus one to bring it back to 0 as the first
				}
				if e.PlayPauseList.Selected == "none" || len(e.PlayPauseList.Selected) < 9 {
					mpirsData.PlayPause = -1
				} else {
					num, _ := strconv.Atoi(e.PlayPauseList.Selected[8:])
					mpirsData.PlayPause = num - 1 // minus one to bring it back to 0 as the first
				}
				if e.NextList.Selected == "none" || len(e.NextList.Selected) < 9 {
					mpirsData.Next = -1
				} else {
					num, _ := strconv.Atoi(e.NextList.Selected[8:])
					mpirsData.Next = num - 1 // minus one to bring it back to 0 as the first
				}

				newControlsApps[e.NameEntry.Text] = mpirsData
			}
		}
		device.AppMediaControls = newControlsApps
		device.SerialNumber = deviceSerial.Text
		cfgGlobal.Devices[name] = device

		cfgGlobal.SaveConfig(configPathGlobal)
		ScanForDevices(a, cfgGlobal, deviceList, connectedDevices, devices, &SerialNumbers)
		w.SetContent(DevicePage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
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
					delete(cfgGlobal.Devices, name)
					ScanForDevices(a, cfgGlobal, deviceList, connectedDevices, devices, &SerialNumbers)
					cfgGlobal.SaveConfig(configPathGlobal)
					w.SetContent(DevicePage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
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
		if SerialNumbers == nil {
			return
		}
		var d dialog.Dialog

		// Create a list of serial numbers for selection
		var serialOptions []string
		for serial := range SerialNumbers {
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
				btn.SetText(fmt.Sprintf("%s - %s", serial, (SerialNumbers)[serial]))
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
			container.NewHBox(widget.NewLabel("Device Name:"), container.New(layout.NewGridWrapLayout(fyne.NewSize(350, deviceName.MinSize().Height)), deviceName)),
			container.NewHBox(widget.NewLabel("Serial Number:"), container.New(layout.NewGridWrapLayout(fyne.NewSize(300, deviceSerial.MinSize().Height)), deviceSerial), searchButton),
		),
	)

	x, _ := processSerialNumber(serialNumberString)
	appVolumeCard := container.NewHBox(layout.NewSpacer())
	masterVolumeCard := container.NewHBox(layout.NewSpacer())
	mediaControlCard := container.NewHBox(layout.NewSpacer())

	if x.numOfSlidders != 0 {
		appVolumeCard = container.NewVBox(
			createSectionHeader("App Volume Configuration"),
			widget.NewSeparator(),
			appVolumeList,
			addAppVolumeButton,
		)

		masterVolumeCard = container.NewVBox(
			createSectionHeader("Master Volume Configuration"),
			widget.NewSeparator(),
			masterVolumeList,
			addMasterVolumeButton,
		)
	}
	if x.numOfButtons != 0 {
		mediaControlCard = container.NewVBox(
			createSectionHeader("Media Configuration"),
			widget.NewSeparator(),
			appControlList,
			addAppControlButton,
		)
	}

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

	var exportimportButton *widget.Button

	if name == "" {
		exportimportButton = importButton
	} else {
		exportimportButton = exportButton
	}

	return container.NewBorder(
		container.NewHBox(backBtn, layout.NewSpacer(), title, layout.NewSpacer(), removeDeviceButton),
		container.NewHBox(exportimportButton, layout.NewSpacer(), saveButton),
		nil,
		nil,
		scrollableList,
	)
}

func SettingsPage(a fyne.App, w fyne.Window, deviceList binding.StringList, connectedDevices binding.BoolList, devices *map[string]*MixMasterInstance, pulseSessions *PulseSessions, mpirsSessions *MpirsSessions) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		cfgGlobal.SaveConfig(configPathGlobal)
		w.SetContent(DevicePage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
	})

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		w.SetContent(DevicePage(a, w, deviceList, connectedDevices, devices, pulseSessions, mpirsSessions))
	})

	startupTitle := canvas.NewText("Startup Settings", color.White)
	startupTitle.TextSize = 24
	startupTitle.TextStyle = fyne.TextStyle{Bold: true}
	startupTitle.Alignment = fyne.TextAlignCenter

	miscTitle := canvas.NewText("Other Settings", color.White)
	miscTitle.TextSize = 24
	miscTitle.TextStyle = fyne.TextStyle{Bold: true}
	miscTitle.Alignment = fyne.TextAlignCenter

	autoStartCheck := widget.NewCheck("Lauch GUI on application Start", func(checked bool) {
		cfgGlobal.App.LaunchGUIOnStart = checked
	})
	autoStartCheck.Checked = cfgGlobal.App.LaunchGUIOnStart

	notificationsEnCheck := widget.NewCheck("Enable Notifications", func(checked bool) {
		cfgGlobal.App.Notifications = checked
	})
	notificationsEnCheck.Checked = cfgGlobal.App.Notifications

	// Wrap in padded, centered, max-width container
	maxWidthContent := container.NewPadded(
		container.NewCenter(
			container.NewVBox(
				startupTitle,
				widget.NewSeparator(),
				autoStartCheck,
				miscTitle,
				widget.NewSeparator(),
				notificationsEnCheck,
			),
		),
	)

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
			container.New(layout.NewGridWrapLayout(fyne.NewSize(200, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			searchButtonPopup(w, "Current Apps Avalible", appNames, entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(290, entry.NumberList.MinSize().Height)), entry.NumberList),
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
			container.New(layout.NewGridWrapLayout(fyne.NewSize(200, entry.NameEntry.MinSize().Height)), entry.NameEntry),
			searchButtonPopup(w, "Current Outputs Avalible", outputNames, entry.NameEntry),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(290, entry.NumberList.MinSize().Height)), entry.NumberList),
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
			container.New(layout.NewGridWrapLayout(fyne.NewSize(110, entry.BackList.MinSize().Height)), entry.BackList),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(110, entry.PlayPauseList.MinSize().Height)), entry.PlayPauseList),
			container.New(layout.NewGridWrapLayout(fyne.NewSize(110, entry.NextList.MinSize().Height)), entry.NextList),
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

func addAppVolumeEntry(w fyne.Window, appName string, appNumber int, appEntries *[]VolumeEntry, appList *fyne.Container, pulseSessions *PulseSessions) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("App Name")
	nameEntry.SetText(appName)

	list := []string{"none"}
	x, err := processSerialNumber(serialNumberString)
	if err != nil {
		return
	}

	for i := range x.numOfSlidders {
		list = append(list, fmt.Sprintf("slidder: %d", i+1))
	}

	numberSelect := widget.NewSelect(list, func(s string) { fmt.Println(s) })

	if appNumber == -1 {
		numberSelect.Selected = "none"
	} else {
		numberSelect.Selected = fmt.Sprintf("Slidder: %d", appNumber+1)
	}

	*appEntries = append(*appEntries, VolumeEntry{nameEntry, numberSelect}) // numberEntry})
	refreshAppVolumeList(w, appEntries, appList, pulseSessions)
}

func addMasterVolumeEntry(w fyne.Window, masterName string, masterNumber int, masterEntries *[]VolumeEntry, appList *fyne.Container, pulseSessions *PulseSessions) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Output Name")
	nameEntry.SetText(masterName)

	list := []string{"none"}
	x, err := processSerialNumber(serialNumberString)
	if err != nil {
		return
	}

	for i := range x.numOfSlidders {
		list = append(list, fmt.Sprintf("slidder: %d", i+1))
	}

	numberSelect := widget.NewSelect(list, func(s string) { fmt.Println(s) })

	if masterNumber == -1 {
		numberSelect.Selected = "none"
	} else {
		numberSelect.Selected = fmt.Sprintf("Slidder: %d", masterNumber+1)
	}

	*masterEntries = append(*masterEntries, VolumeEntry{nameEntry, numberSelect})
	refreshMasterVolumeList(w, masterEntries, appList, pulseSessions)
}

func addAppControlEntry(w fyne.Window, appName string, controlNumbers *mpirsData, appControlEntries *[]AppControlEntry, appControlList *fyne.Container, mpirsSessions *MpirsSessions) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("App Name")
	nameEntry.SetText(appName)

	list := []string{"none"}
	x, err := processSerialNumber(serialNumberString)
	if err != nil {
		return
	}

	for i := range x.numOfButtons {
		list = append(list, fmt.Sprintf("button: %d", i+1))
	}

	// Back
	backList := widget.NewSelect(list, func(s string) { fmt.Println(s) })
	// PlayPause
	playpauseList := widget.NewSelect(list, func(s string) { fmt.Println(s) })
	// Next
	nextList := widget.NewSelect(list, func(s string) { fmt.Println(s) })

	if controlNumbers != nil {
		if controlNumbers.Back == -1 {
			backList.Selected = "none"
		} else {
			backList.Selected = fmt.Sprintf("Button: %d", controlNumbers.Back+1)
		}
		if controlNumbers.PlayPause == -1 {
			playpauseList.Selected = "none"
		} else {
			playpauseList.Selected = fmt.Sprintf("Button: %d", controlNumbers.PlayPause+1)
		}
		if controlNumbers.Next == -1 {
			nextList.Selected = "none"
		} else {
			nextList.Selected = fmt.Sprintf("Button: %d", controlNumbers.Next+1)
		}
	} else {
		backList.Selected = "none"
		playpauseList.Selected = "none"
		nextList.Selected = "none"
	}

	*appControlEntries = append(*appControlEntries, AppControlEntry{nameEntry, backList, playpauseList, nextList})
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

	if len(serialNumber) != 13 {
		return ControlsCount{}, fmt.Errorf("serial number is the wrong length: %d", len(serialNumber))
	}

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
