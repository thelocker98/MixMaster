// internal/ui/devices.go
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DeviceGUI struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Connected bool              `json:"connected"`
	Apps      []*App            `json:"apps,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"` // optional extra info
}

func NewDeviceListView(w fyne.Window) fyne.CanvasObject {
	devices := []*DeviceGUI{
		{ID: "1", Name: "Mixer 1", Connected: true},
		{ID: "2", Name: "Light Controller", Connected: false},
	}

	list := widget.NewList(
		func() int { return len(devices) },
		func() fyne.CanvasObject { return widget.NewButton("", nil) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			dev := devices[i]
			o.(*widget.Button).SetText(fmt.Sprintf("%s (%v)", dev.Name, onlineStatus(dev.Connected)))
			o.(*widget.Button).OnTapped = func() {
				editView := NewDeviceEditView(w, dev)
				w.SetContent(editView)
			}
		},
	)

	addBtn := widget.NewButton("Add Device", func() {
		fmt.Println("Add device clicked")
	})

	return container.NewBorder(nil, addBtn, nil, nil, list)
}

func onlineStatus(ok bool) string {
	if ok {
		return "Online"
	}
	return "Offline"
}
