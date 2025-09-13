package izapple2

import (
	"fmt"
)

/*
Profile card for Apple II Interface.

The firmware source code for 5MB is available at: https://bitsavers.org/pdf/apple/disk/profile/appleII_interface/AII_Profile_Boot_Prom_198402.pdf
See also:
	https://bitsavers.org/pdf/apple/disk/profile/072-0116_Profile_Level_II_Service_Manual_Oct84.pdf
	https://bitsavers.org/pdf/apple/disk/profile/Profile_Communication_Protocol.pdf
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Hard%20Disk%20Drive%20Controllers/Apple%20II%20ProFile%20Interface%20Card/

Softswitches:
  W0: WR_PORT, write to Z8 RAM (byte at a time)
  R1: RD_PORT, read from Z8 RAM (byte at a time)
  R2: BUSY, Z8 not ready
	  bit 0 off is drive present and connected to card
	  bit 6 is parity error
	  bit 7 on is ready and online
  W2: RSTAT?
  W3: CLR PARITY, clear any previous parity error

*/

// CardProfile represents the Apple II Profile interface card
type CardProfile struct {
	cardBase
}

func newCardProfileBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Apple II Profile interface card",
		description: "Profile interface card - not implemented",
		hide:        true,
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/Apple II Interface 341-0271-A ROM 2716 for 5MB Profile.bin"},
			//{"rom", "ROM file to load", "<internal>/Apple II Interface 341-0299-B ROM 2716 for 10MB Profile.bin"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardProfile
			romFile := paramsGetPath(params, "rom")
			err := c.loadRomFromResource(romFile, cardRomUpperEnd)
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
	}
}

func (c *CardProfile) assign(a *Apple2, slot int) {

	// WR_PORT, write to Z8 RAM (byte at a time)
	c.addCardSoftSwitchW(0, func(value uint8) {
		fmt.Printf("[cardProfile] WR_PORT: Write to Z8 RAM, value 0x%02x\n", value)
	}, "PROFILE_WR_PORT")

	// RD_PORT, read from Z8 RAM (byte at a time)
	c.addCardSoftSwitchR(1, func() uint8 {
		fmt.Printf("[cardProfile] RD_PORT: Read from Z8 RAM\n")
		return 0x00 // Return dummy data for now
	}, "PROFILE_RD_PORT")

	// BUSY, Z8 not ready
	c.addCardSoftSwitchR(2, func() uint8 {
		fmt.Printf("[cardProfile] BUSY: Z8 status check\n")
		// Return 0x80 (bit 7 on) to indicate ready and online
		// bit 0 off indicates drive present
		return 0x80
	}, "PROFILE_BUSY")

	// RSTAT?
	c.addCardSoftSwitchW(2, func(value uint8) {
		fmt.Printf("[cardProfile] RSTAT: Write status, value 0x%02x\n", value)
	}, "PROFILE_RSTAT")

	// CLR PARITY, clear any previous parity error
	c.addCardSoftSwitchW(3, func(value uint8) {
		fmt.Printf("[cardProfile] CLR_PARITY: Clear parity error, value 0x%02x\n", value)
	}, "PROFILE_CLR_PARITY")

	// Call base assignment to setup ROM and other card functionality
	c.cardBase.assign(a, slot)
}
