package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

type panelDevices struct {
	s        *state
	w        fyne.Widget
	joystick *panelJoystick
}

func newPanelDevices(s *state) *panelDevices {
	var pd panelDevices
	pd.s = s

	pd.joystick = newPanelJoystick()

	pd.w = widget.NewVBox(pd.joystick.w)

	return &pd
}
