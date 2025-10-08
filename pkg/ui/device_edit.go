// internal/ui/device_edit.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewDeviceEditView(w fyne.Window, dev *DeviceGUI) fyne.CanvasObject {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(dev.Name)

	appsList := widget.NewList(
		func() int { return len(dev.Apps) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(dev.Apps[i].Name)
		},
	)

	backBtn := widget.NewButton("Back", func() {
		w.SetContent(NewDeviceListView(w))
	})

	addAppBtn := widget.NewButton("Add App", func() {
		w.SetContent(NewAppSelectView(w, dev))
	})

	return container.NewBorder(backBtn, addAppBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Edit Device"),
			widget.NewForm(widget.NewFormItem("Name", nameEntry)),
			widget.NewLabel("Assigned Apps:"),
			appsList,
		),
	)
}
