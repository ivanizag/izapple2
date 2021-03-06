package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ivanizag/izapple2/screen"
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
		theme.NewThemedResource(resourceTelevisionClassicSvg),
		func() {
			tbs.setScreenMode(screen.ScreenModeNTSC)
		})
	tbs.ntscDisabled = widget.NewIcon(
		theme.NewDisabledResource(resourceTelevisionClassicSvg))

	tbs.plain = widget.NewButtonWithIcon("",
		theme.NewThemedResource(resourceTelevisionSvg),
		func() {
			tbs.setScreenMode(screen.ScreenModePlain)
		})
	tbs.plainDisabled = widget.NewIcon(
		theme.NewDisabledResource(resourceTelevisionSvg))

	tbs.green = widget.NewButtonWithIcon("",
		theme.NewThemedResource(resourceMonitorSvg),
		func() {
			tbs.setScreenMode(screen.ScreenModeGreen)
		})
	tbs.greenDisabled = widget.NewIcon(
		theme.NewDisabledResource(resourceMonitorSvg))

	tbs.w = container.NewHBox(
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
	case screen.ScreenModeNTSC:
		tbs.ntsc.Show()
		tbs.ntscDisabled.Hide()
	case screen.ScreenModePlain:
		tbs.plain.Show()
		tbs.plainDisabled.Hide()
	case screen.ScreenModeGreen:
		tbs.green.Show()
		tbs.greenDisabled.Hide()
	}

	tbs.s.screenMode = screenMode

	switch screenMode {
	case screen.ScreenModeNTSC:
		tbs.ntsc.Hide()
		tbs.ntscDisabled.Show()
	case screen.ScreenModePlain:
		tbs.plain.Hide()
		tbs.plainDisabled.Show()
	case screen.ScreenModeGreen:
		tbs.green.Hide()
		tbs.greenDisabled.Show()
	}
}

func (tbs *toolbarScreen) ToolbarObject() fyne.CanvasObject {
	return tbs.w
}
