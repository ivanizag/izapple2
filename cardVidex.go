package izapple2

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/ivanizag/izapple2/component"
)

/*
	Videx 80 columns card for the Apple II+

See:
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/80%20Column%20Cards/Videx%20Videoterm/Manuals/Videx%20Videoterm%20-%20Installation%20and%20Operation%20Manual.pdf
	http://bitsavers.trailing-edge.com/components/motorola/_dataSheets/6845.pdf
	https://glasstty.com/?p=660

*/

// CardVidex represents a Videx compatible 80 column card
type CardVidex struct {
	cardBase
	mc6845   component.MC6845
	sramPage uint8
	sram     [0x800]uint8
	upperROM memoryHandler
	charGen  []uint8
}

func newCardVidexBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Videx 80 columns Card",
		description: "Videx compatible 80 columns card",
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/Videx Videoterm ROM 2.4.bin"},
			{"charmap", "Character map file to load", "<internal>/80ColumnP110.BIN"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardVidex

			// The C800 area has ROM and RAM
			err := c.loadRomFromResource("<internal>/Videx Videoterm ROM 2.4.bin")
			if err != nil {
				return nil, err
			}
			c.upperROM = c.romC8xx
			c.romC8xx = &c

			err = c.loadCharacterMap(paramsGetPath(params, "charmap"))
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
	}
}

func (c *CardVidex) loadCharacterMap(filename string) error {
	bytes, _, err := LoadResource(filename)
	if err != nil {
		return err
	}
	size := len(bytes)
	if size < 0x800 {
		return errors.New("character ROM size not supported for Videx")
	}
	c.charGen = bytes
	return nil
}

func (c *CardVidex) assign(a *Apple2, slot int) {

	// TODO: use addCardSoftSwitches()
	for i := uint8(0x0); i <= 0xf; i++ {
		// Bit 0 goes to the RS pin of the MC6548. It controls
		// whether a register is being accesed or the contents
		// of the register is being accessed
		rsPin := (i & 1) == 1

		// Bits 2 and 3 determine which page will be selected
		sramPage := i >> 2

		ssName := fmt.Sprintf("VIDEXPAGE%v", sramPage)
		if rsPin {
			ssName += "REG"
		} else {
			ssName += "ADDRESS"
		}

		c.addCardSoftSwitchR(i, func() uint8 {
			c.sramPage = sramPage
			return c.mc6845.Read(rsPin)
		}, ssName+"R")
		c.addCardSoftSwitchW(i, func(value uint8) {
			c.sramPage = sramPage
			c.mc6845.Write(rsPin, value)
		}, ssName+"W")
	}

	c.cardBase.assign(a, slot)
	a.softVideoSwitch = NewSoftVideoSwitch(c)
}

const videxRomLimit = uint16(0xcc00)
const videxSramLimit = uint16(0xce00)
const videxSramMask = uint16(0x01ff)

func (c *CardVidex) peek(address uint16) uint8 {
	if address < videxRomLimit {
		return c.upperROM.peek(address)
	} else if address < videxSramLimit {
		return c.sram[address&videxSramMask+uint16(c.sramPage)*0x200]
	}
	return 0
}

func (c *CardVidex) poke(address uint16, value uint8) {
	if address >= videxRomLimit && address < videxSramLimit {
		c.sram[address&videxSramMask+uint16(c.sramPage)*0x200] = value
	}
}

func (c *CardVidex) setBase(base uint16) {
	// Nothing
}

const (
	videxCharWidth = uint8(8)
)

func (c *CardVidex) buildImage(light color.Color) *image.RGBA {
	params := c.mc6845.ImageData()
	width, height := params.DisplayedWidthHeight(videxCharWidth)
	if (width == 0) || (height == 0) {
		// No image available
		size := image.Rect(0, 0, 3, 3)
		img := image.NewRGBA(size)
		img.Set(1, 1, color.White)
		return img
	}
	ms := time.Now().Nanosecond() / (1000 * 1000) // Host time, used for the cursoR blink

	size := image.Rect(0, 0, width, height)
	img := image.NewRGBA(size)

	params.IterateScreen(func(address uint16, charLine uint8,
		cursorMode uint8, displayEnable bool,
		column uint8, y int) {

		bits := uint8(0)
		if displayEnable {
			char := c.sram[address&0x7ff]
			bits = c.charGen[(uint16(char&0x7f)<<4)+uint16(charLine)]
			isCursor := false
			switch cursorMode {
			case component.MC6845CursorFixed:
				isCursor = true
			case component.MC6845CursorSlow:
				// It should be 533ms (32/60, 32 screen refreshes)
				// Let's make a 2 blinks per second
				isCursor = ms/2 > 1000/4
			case component.MC6845CursorFast:
				// It should be 266ms (32/60, 16 screen refreshes)
				// Let's make a 4 blinks per second
				isCursor = ms/4 > 1000/8
			}
			if isCursor {
				bits = ^bits
			}
			if char >= 128 {
				// Inverse
				bits = ^bits
			}
		}

		x := int(column) * int(videxCharWidth)

		for i := 0; i < int(videxCharWidth); i++ {
			pixel := (bits & 0x80) != 0
			if pixel {
				img.Set(x, y, light)
			} else {
				img.Set(x, y, color.Black)
			}
			bits <<= 1
			x++
		}
	})

	return img
}
