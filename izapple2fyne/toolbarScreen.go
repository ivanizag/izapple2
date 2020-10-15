package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/ivanizag/izapple2"
)

type toolbarScreen struct {
	s             *state
	w             fyne.CanvasObject
	ntsc          *widget.Button
	ntscDisabled  *widget.Icon
	plain         *widget.Button
	plainDisabled *widget.Icon
	green         *widget.Button
	greenDisabled *widget.Icon
}

func newToolbarScreen(s *state) *toolbarScreen {
	var tbs toolbarScreen
	tbs.s = s

	tbs.ntsc = widget.NewButtonWithIcon("",
		theme.NewThemedResource(resourceTelevisionClassicSvg, nil),
		func() {
			tbs.setScreenMode(izapple2.ScreenModeNTSC)
		})
	tbs.ntscDisabled = widget.NewIcon(
		theme.NewDisabledResource(resourceTelevisionClassicSvg))

	tbs.plain = widget.NewButtonWithIcon("",
		theme.NewThemedResource(resourceTelevisionSvg, nil),
		func() {
			tbs.setScreenMode(izapple2.ScreenModePlain)
		})
	tbs.plainDisabled = widget.NewIcon(
		theme.NewDisabledResource(resourceTelevisionSvg))

	tbs.green = widget.NewButtonWithIcon("",
		theme.NewThemedResource(resourceMonitorSvg, nil),
		func() {
			tbs.setScreenMode(izapple2.ScreenModeGreen)
		})
	tbs.greenDisabled = widget.NewIcon(
		theme.NewDisabledResource(resourceMonitorSvg))

	tbs.w = widget.NewHBox(
		tbs.ntsc, tbs.ntscDisabled,
		tbs.plain, tbs.plainDisabled,
		tbs.green, tbs.greenDisabled)

	tbs.ntscDisabled.Hide()
	tbs.plainDisabled.Hide()
	tbs.greenDisabled.Hide()
	tbs.setScreenMode(s.screenMode)

	return &tbs
}

func (tbs *toolbarScreen) setScreenMode(screenMode int) {
	switch tbs.s.screenMode {
	case izapple2.ScreenModeNTSC:
		tbs.ntsc.Show()
		tbs.ntscDisabled.Hide()
	case izapple2.ScreenModePlain:
		tbs.plain.Show()
		tbs.plainDisabled.Hide()
	case izapple2.ScreenModeGreen:
		tbs.green.Show()
		tbs.greenDisabled.Hide()
	}

	tbs.s.screenMode = screenMode

	switch screenMode {
	case izapple2.ScreenModeNTSC:
		tbs.ntsc.Hide()
		tbs.ntscDisabled.Show()
	case izapple2.ScreenModePlain:
		tbs.plain.Hide()
		tbs.plainDisabled.Show()
	case izapple2.ScreenModeGreen:
		tbs.green.Hide()
		tbs.greenDisabled.Show()
	}
}

func (tbs *toolbarScreen) ToolbarObject() fyne.CanvasObject {
	return tbs.w
}
