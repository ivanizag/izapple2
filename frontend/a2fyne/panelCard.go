package main

import (
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/ivanizag/izapple2"
)

type panelCard struct {
	w fyne.Widget
}

func newPanelCard(slot int, card izapple2.Card) *panelCard {
	var pc panelCard

	form := widget.NewForm()
	form.Append("slot", widget.NewLabel(strconv.Itoa(slot)))

	info := card.GetInfo()
	if info != nil {
		for k, v := range info {
			form.Append(k, widget.NewLabel(v))
		}
	}

	pc.w = widget.NewGroup(card.GetName(), form)
	return &pc
}
