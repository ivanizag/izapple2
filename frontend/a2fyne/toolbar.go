package main

import (
	"fmt"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func buildCommandToolbar(s *state, icon fyne.Resource, command int) widget.ToolbarItem {
	return widget.NewToolbarAction(
		theme.NewThemedResource(icon, nil), func() {
			s.a.SendCommand(command)
		})
}

func buildToolbar(s *state) *widget.Toolbar {
	tb := widget.NewToolbar()
	tb.Append(buildCommandToolbar(s, resourceRestartSvg, izapple2.CommandReset))
	tb.Append(buildCommandToolbar(s, resourcePauseSvg, izapple2.CommandPauseUnpauseEmulator))
	tb.Append(buildCommandToolbar(s, resourceFastForwardSvg, izapple2.CommandToggleSpeed))
	tb.Append(widget.NewToolbarSeparator())
	tb.Append(newToolbarScreen(s))
	tb.Append(widget.NewToolbarSeparator())
	tb.Append(widget.NewToolbarAction(
		theme.NewThemedResource(resourceLayersTripleSvg, nil), func() {
			s.showPages = !s.showPages
			if !s.showPages {
				s.win.SetTitle(s.DefaultTitle())
			}
		}))
	tb.Append(widget.NewToolbarAction(
		theme.NewThemedResource(resourceCameraSvg, nil), func() {
			err := screen.SaveSnapshot(s.a, s.screenMode, "snapshot.png")
			if err != nil {
				s.app.SendNotification(fyne.NewNotification(
					s.win.Title(),
					fmt.Sprintf("Error saving snapshoot: %v.\n.", err)))
			} else {
				s.app.SendNotification(fyne.NewNotification(
					s.win.Title(),
					"Snapshot saved on 'snapshot.png'"))
			}
		}))
	tb.Append(widget.NewToolbarSeparator())
	tb.Append(widget.NewToolbarAction(
		theme.NewThemedResource(resourceCapsLockSvg, nil), func() {
			s.a.SetForceCaps(!s.a.IsForceCaps())
			s.win.SetTitle(s.DefaultTitle())
		}))
	//tb.Append(widget.NewToolbarSeparator())
	//tb.Append(newToolbarDisk("S6D1")
	tb.Append(widget.NewToolbarSpacer())
	tb.Append(widget.NewToolbarAction(
		theme.ViewFullScreenIcon(),
		func() {
			s.win.SetFullScreen(!s.win.FullScreen())
		}))
	tb.Append(widget.NewToolbarAction(
		theme.NewThemedResource(resourcePageLayoutSidebarRightSvg, nil),
		func() {
			w := s.devices.w
			if w.Visible() {
				w.Hide()
			} else {
				w.Show()
			}
		}))

	return tb
}
