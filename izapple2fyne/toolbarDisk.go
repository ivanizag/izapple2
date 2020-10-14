package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
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
	return widget.NewHBox(
		widget.NewIcon(theme.VolumeUpIcon()),
		widget.NewLabel(tbd.name),
		widget.NewLabel("track 12"),
		widget.NewButton("eject", nil),
	)
}
