package main

import (
	"fyne.io/fyne/widget"
)

type panelDevices struct {
	s        *state
	w        *widget.Box
	joystick *panelJoystick
}

func newPanelDevices(s *state) *panelDevices {
	var pd panelDevices
	pd.s = s
	pd.w = widget.NewVBox()

	pd.joystick = newPanelJoystick()
	pd.w.Append(pd.joystick.w)

	var cards = s.a.GetCards()
	for i, card := range cards {
		if card != nil && card.GetName() != "" {
			pd.w.Append(newPanelCard(i, card).w)
		}
	}

	return &pd
}
