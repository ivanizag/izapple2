package izapple2

/*
Simulates just what is needed to make Total Replay use fast mode. Can change
from controlled speed to max speed the emulator can do.
Note: It ends up not being useful for Total Replay as loading from HD is already
very fast. HD blocks are loaded directly on the emulated RAM.

Note that it doesn't intefere with the Apple IIe 80 columns in slot 3. It doesn't
have ROM or slot specific sofswitches.


See:
	https://github.com/a2-4am/4cade/blob/master/src/hw.accel.a
	http://www.a2heaven.com/webshop/resources/pdf_document/18/82/c.pdf

*/

// CardFastChip represents a
type CardFastChip struct {
	cardBase
	unlocked       bool
	unlockCounter  uint8
	enabled        bool
	accelerated    bool
	configRegister uint8
}

func newCardFastChipBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "FASTChip IIe Card - limited",
		description: "Accelerator card for Apple IIe (limited support)",
		buildFunc: func(params map[string]string) (Card, error) {
			return &CardFastChip{}, nil
		},
	}
}

const (
	fastChipUnlockToken   = 0x6a
	fastChipUnlockRepeats = 4
	fastChipNormalSpeed   = uint8(9)
)

func (c *CardFastChip) assign(a *Apple2, slot int) {
	// The softswitches are outside the card reserved ss
	// Only writes are implemented to avoid conflicts with the joysticks
	a.io.addSoftSwitchW(0x6a, func(value uint8) {
		if value == fastChipUnlockToken {
			c.unlockCounter++
			if c.unlockCounter >= fastChipUnlockRepeats {
				c.unlocked = true
			}
		} else {
			c.unlockCounter = 0
			c.unlocked = false
			c.enabled = false
		}
	}, "FASTCHIP-LOCK")

	a.io.addSoftSwitchW(0x6b, func(uint8) {
		if c.unlocked {
			c.enabled = true
		}
	}, "FASTCHIP-ENABLE")

	a.io.addSoftSwitchW(0x6d, func(value uint8) {
		if c.enabled {
			c.setSpeed(a, value)
		}
	}, "FASTCHIP-SPEED")

	a.io.addSoftSwitchW(0x6e, func(value uint8) {
		if c.enabled {
			c.configRegister = value
		}
	}, "FASTCHIP-CONFIG")

	a.io.addSoftSwitchW(0x6f, func(value uint8) {
		if c.enabled && c.configRegister == 0 {
			c.setSpeed(a, value)
		}
	}, "FASTCHIP-CONFIG")

	c.cardBase.assign(a, slot)
}

func (c *CardFastChip) setSpeed(a *Apple2, value uint8) {
	newAccelerated := (value > fastChipNormalSpeed)
	if newAccelerated == c.accelerated {
		// No change requested
		return
	}
	if newAccelerated {
		a.RequestFastMode()
	} else {
		a.ReleaseFastMode()
	}
	c.accelerated = newAccelerated
}
