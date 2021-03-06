package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type toolbarDisk struct {
	name string
}

func newToolbarDisk(name string) *toolbarDisk {
	var tbd toolbarDisk
	tbd.name = name
	return &tbd
}

func (tbd *toolbarDisk) ToolbarObject() fyne.CanvasObject {
	return container.NewHBox(
		widget.NewIcon(theme.VolumeUpIcon()),
		widget.NewLabel(tbd.name),
		widget.NewLabel("track 12"),
		widget.NewButton("eject", nil),
	)
}
