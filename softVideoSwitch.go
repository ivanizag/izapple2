package izapple2

import (
	"image"
	"image/color"
)

/*
	Videx Soft Video Switch external on the Videx Videoterm and integrated on the Videx Ultraterm

	See:
		https://archive.org/details/videx-soft-video-switch
*/

type softVideoSwitch interface {
	buildImage(light color.Color) *image.RGBA
	isSoftSwitchActive() bool
}

func (a *Apple2) setSoftVideoSwitch(card softVideoSwitch) {
	a.softVideoSwitch = card
}

func (a *Apple2) isSoftVideoSwitchActive() bool {
	if a.softVideoSwitch == nil {
		return false
	}

	return a.softVideoSwitch.isSoftSwitchActive()
}
