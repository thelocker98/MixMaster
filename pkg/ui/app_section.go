// internal/ui/app_select.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	ID   int64
	Name string
}

func NewAppSelectView(w fyne.Window, dev *DeviceGUI) fyne.CanvasObject {
	available := []*App{
		{ID: 1, Name: "Audio Mixer"},
		{ID: 2, Name: "Light Show"},
		{ID: 3, Name: "Sensor Monitor"},
	}

	list := widget.NewList(
		func() int { return len(available) },
		func() fyne.CanvasObject { return widget.NewButton("", nil) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			app := available[i]
			o.(*widget.Button).SetText(app.Name)
			o.(*widget.Button).OnTapped = func() {
				dev.Apps = append(dev.Apps, app)
				w.SetContent(NewDeviceEditView(w, dev))
			}
		},
	)

	back := widget.NewButton("Cancel", func() {
		w.SetContent(NewDeviceEditView(w, dev))
	})

	return container.NewBorder(nil, back, nil, nil, list)
}
