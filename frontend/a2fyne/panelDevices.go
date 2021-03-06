package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type panelDevices struct {
	s        *state
	w        *fyne.Container
	joystick *panelJoystick
}

func newPanelDevices(s *state) *panelDevices {
	var pd panelDevices
	pd.s = s
	pd.w = container.NewVBox()

	pd.joystick = newPanelJoystick()
	pd.w.Add(pd.joystick.w)

	var cards = s.a.GetCards()
	for i, card := range cards {
		if card != nil && card.GetName() != "" {
			pd.w.Add(newPanelCard(i, card).w)
		}
	}

	return &pd
}
