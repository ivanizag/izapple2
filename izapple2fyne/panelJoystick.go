package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

type panelJoystick struct {
	w         fyne.Widget
	labelJoy1 *widget.Label
	labelJoy2 *widget.Label
}

const textJoystickNotAvailable = "unplugged"

func newPanelJoystick() *panelJoystick {
	var pj panelJoystick

	pj.labelJoy1 = widget.NewLabel("")
	pj.labelJoy2 = widget.NewLabel("")
	widget.NewForm()
	pj.w = widget.NewGroup(
		"Joysticks",
		widget.NewForm(
			widget.NewFormItem("Joystick 1", pj.labelJoy1),
			widget.NewFormItem("Joystick 2", pj.labelJoy2),
		),
	)

	return &pj
}

func (pj *panelJoystick) updateJoy1(info *joystickInfo) {
	newName := textJoystickNotAvailable
	if info != nil {
		newName = info.name
	}
	if newName != pj.labelJoy1.Text {
		pj.labelJoy1.SetText(newName)
	}
}

func (pj *panelJoystick) updateJoy2(info *joystickInfo) {
	newName := textJoystickNotAvailable
	if info != nil {
		newName = info.name
	}
	if newName != pj.labelJoy2.Text {
		pj.labelJoy2.SetText(newName)
	}
}
