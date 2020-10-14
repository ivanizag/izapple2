package izapple2

import (
	"fmt"
)

/*
Logger card. It never existed, I use it to trace accesses to the card.
*/

// CardLogger is a fake card to log soft switch invocations
type CardLogger struct {
	cardBase
}

// NewCardLogger creates a new VidHD card
func NewCardLogger() *CardLogger {
	var c CardLogger
	c.name = "Softswitch log card"
	return &c
}

func (c *CardLogger) assign(a *Apple2, slot int) {
	for i := uint8(0x0); i <= 0xf; i++ {
		iCopy := i
		c.addCardSoftSwitchR(i, func(*ioC0Page) uint8 {
			fmt.Printf("[cardLogger] Read access to softswith 0x%x for slot %v.\n", iCopy, slot)
			return 0
		}, "LOGGERR")
		c.addCardSoftSwitchW(i, func(_ *ioC0Page, value uint8) {
			fmt.Printf("[cardLogger] Write access to softswith 0x%x for slot %v, value 0x%v.\n", iCopy, slot, value)
		}, "LOGGERW")
	}

	c.cardBase.assign(a, slot)
}
