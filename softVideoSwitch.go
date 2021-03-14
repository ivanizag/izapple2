package izapple2

import (
	"image"
	"image/color"
)

/*
	Videx Soft Video Switch

	See:
		https://archive.org/details/videx-soft-video-switch

*/

// SoftVideoSwitch represents a Videx soft video switch
type SoftVideoSwitch struct {
	card *CardVidex
}

// NewSoftVideoSwitch creates a new SoftVideoSwitch
func NewSoftVideoSwitch(card *CardVidex) *SoftVideoSwitch {
	var vs SoftVideoSwitch
	vs.card = card
	return &vs
}

func (vs *SoftVideoSwitch) isActive() bool {
	if vs == nil {
		return false
	}

	isTextMode := vs.card.a.io.isSoftSwitchActive(ioFlagText)
	ann0 := vs.card.a.io.isSoftSwitchActive(ioFlagAnnunciator0)
	return isTextMode && ann0
}

func (vs *SoftVideoSwitch) BuildAlternateImage(light color.Color) *image.RGBA {
	return vs.card.buildImage(light)
}

func (a *Apple2) SoftVideoSwitch() *SoftVideoSwitch {
	return a.softVideoSwitch
}
