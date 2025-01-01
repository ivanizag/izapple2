package izapple2

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	"github.com/ivanizag/izapple2/component"
)

/*
	Videx Ultraterm 80 columns card for the Apple II+

See:
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/80%20Column%20Cards/Videx%20UltraTerm/
	http://www.bitsavers.org/components/motorola/_dataSheets/6845.pdf


Firmware:
    The firmware to operate your UltraTerm is contained in a 4K-byte 2732A
EPROM, U6. The lower half of this IC contains seven versions of the code
which appears at $CN00 to $CNFF, one for each slot.
There are 2K bytes of address space available for use in the co-resident
memory space at $C800. However, the upper 1 K bytes of this space is
used by the video refresh memory. For this reason the firmware is split into
two banks. These banks are selected with bit seven of the MCP When the
second bank of firmware is selected it overlays the Video Refresh Memory
(VRM) at addresses from $CC00 to $CFE0. The first bank of the firmware
always occupies the region from $C800 to $CBFF.

Formats (from the firmware listing, appendix F of the manual):
	0: 80 x 24 non-interlaced (low density chars) (sram512, Apple Video source ??)
	1: 96 x 24 non-interlaced (low density chars) (sram256, 17 Mhz)
	2: 160 x 24 non-interlaced (low density chars) (sram256, 28 mhz)
	3: 80 x 24 interlaced (high density chars) (sram256, 17 Mhz)
	4: 80 x 32 interlaced (high density chars) (sram256, 17 Mhz)
	5: 80 x 48 interlaced (low density chars) (sram256, 17 Mhz)
	6: 160 x 24 interlaced (used for 132 x 24 screen) (high density chars) (sram256, 28 mhz)
	7: 128 X 32 interlaced  (high density chars) (sram256, 28 mhz)
*/

// CardVidex represents a Videx Ultraterm compatible 80 column card
type CardVidexUltraterm struct {
	cardBase
	mc6845         component.MC6845
	modeControl    uint8
	videoAttribute uint8
	cxrom          memoryHandler
	alwaysShow     bool
	sramPage512    uint8 // sram page on 512 kb mode (videoterm emultation)

	sram    [0x1000]uint8
	charGen []uint8
}

func newCardVidexUltratermBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Videx Ultraterm 80 columns Card",
		description: "Videx Utraterm compatible 80 columns card",
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/videx_ultraterm_frm_b537.bin"},
			{"charmap", "Character map file to load", "<internal>/videx_ultraterm_chs_7859.bin"},
			{"always", "Always show the 80 columns output", "false"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardVidexUltraterm

			// The C800 area has ROM and RAM
			err := c.loadRomFromResource(paramsGetPath(params, "rom"), cardRomFull)
			if err != nil {
				return nil, err
			}
			c.cxrom = c.romCxxx
			c.romCxxx = &c

			err = c.loadCharacterMap(paramsGetPath(params, "charmap"))
			if err != nil {
				return nil, err
			}

			c.alwaysShow = paramsGetBool(params, "always")
			c.videoAttribute = videxUltratermDefaultVideoAttribute
			return &c, nil
		},
	}
}

func (c *CardVidexUltraterm) loadCharacterMap(filename string) error {
	bytes, _, err := LoadResource(filename)
	if err != nil {
		return err
	}
	size := len(bytes)
	if size < 0x1000 {
		return errors.New("character ROM size not supported for Videx")
	}
	c.charGen = bytes
	return nil
}

func (c *CardVidexUltraterm) assign(a *Apple2, slot int) {

	for page := uint8(0); page < 4; page++ {
		bitsA3A2 := page << 2
		ssName := fmt.Sprintf("ULTRATERMPAGE%v", page)

		pageCopy := page
		c.addCardSoftSwitchR(0+bitsA3A2, func() uint8 {
			c.sramPage512 = pageCopy
			return c.mc6845.Read(false)
		}, ssName+"REGR")
		c.addCardSoftSwitchW(0+bitsA3A2, func(value uint8) {
			c.mc6845.Write(false, value)
		}, ssName+"REGW")

		c.addCardSoftSwitchR(1+bitsA3A2, func() uint8 {
			c.sramPage512 = pageCopy
			return c.mc6845.Read(true)
		}, ssName+"VALR")
		c.addCardSoftSwitchW(1+bitsA3A2, func(value uint8) {
			c.mc6845.Write(true, value)
		}, ssName+"VALW")

		c.addCardSoftSwitchR(2+bitsA3A2, func() uint8 {
			c.sramPage512 = pageCopy
			return c.modeControl
		}, ssName+"MODER")
		c.addCardSoftSwitchW(2+bitsA3A2, func(value uint8) {
			c.modeControl = value
		}, ssName+"MODEW")

		c.addCardSoftSwitchR(3+bitsA3A2, func() uint8 {
			c.sramPage512 = pageCopy
			return c.videoAttribute
		}, ssName+"ATTRE")
		c.addCardSoftSwitchW(3+bitsA3A2, func(value uint8) {
			c.videoAttribute = value
		}, ssName+"ATTRW")
	}

	c.cardBase.assign(a, slot)
	a.setSoftVideoSwitch(c)
}

/*
Bit:

	7 Firmware Page Select
	6 Video Signal Select
		1 = UltraTerm
	5 Clock Frequency
		1 = 28.7595,0 = 17.430 MHz
	4 Character Address Format
		1 = 256-Byte Pages, 0 = 512-Byte Blocks
	3 Character RAM Address bit 11 (256-byte mode)
	2 Character RAM Address bit 10 (256-byte mode)
	1 Character RAM Address bit 9 (256-byte mode)
	0 Character RAM Address bit 8 (256-by1e mode)
*/
const (
	videxUltratermMCPFirmwarePageSelect = uint8(0x80)
	videxUltratermMCPVideoSignalSelect  = uint8(0x40)
	videxUltratermMCPClockFrequency     = uint8(0x20)
	videxUltratermMCPSRamAdressFormat   = uint8(0x10)
	videxUltratermMCPSramPageMask       = uint8(0x0f)
)

const (
	videxUltratermAttributesHighlight     = uint8(1) // Lowlight or Highlight
	videxUltratermAttributesInverse       = uint8(2) // Normal or Inverse
	videxUltratermAttributesAlternateChar = uint8(4) // Normal or Alternate (low quality) character sets
)

const videxUltratermDefaultVideoAttribute = uint8(0x00 | videxUltratermAttributesInverse<<4) // LowlightNormal and LowlightInverse

const videxUltratermSramStart = uint16(0xcc00)
const videxUltratermSramLegacyMask = uint16(0x01ff)
const videxUltratermSramMask = uint16(0x0ff)
const videxUltratermSram512Mask = uint16(0x7ff)
const videxUltratermSram256Mask = uint16(0xfff)

func (c *CardVidexUltraterm) sramAddress(address uint16) uint16 {
	is512mode := c.modeControl&videxUltratermMCPSRamAdressFormat == 0
	if is512mode {
		// Legacy or 512 mode
		return address&videxUltratermSramLegacyMask + uint16(c.sramPage512)*512
	}

	sramPage256 := c.modeControl & videxUltratermMCPSramPageMask
	return address&videxUltratermSramMask + uint16(sramPage256)*256
}

func (c *CardVidexUltraterm) peek(address uint16) uint8 {
	isFirmwarePageSelected := c.modeControl&videxUltratermMCPFirmwarePageSelect != 0
	if address < videxUltratermSramStart || isFirmwarePageSelected {
		return c.cxrom.peek(address)
	}
	return c.sram[c.sramAddress(address)]
}

func (c *CardVidexUltraterm) poke(address uint16, value uint8) {
	if address >= videxUltratermSramStart {
		c.sram[c.sramAddress(address)] = value
	}
}

func (c *CardVidexUltraterm) isSoftSwitchActive() bool {
	if c.alwaysShow {
		return true
	}
	return c.modeControl&videxUltratermMCPVideoSignalSelect != 0
}

const (
	videxUltratermCharWidth = uint8(9)
)

func (c *CardVidexUltraterm) colorsPerAttributes(topBit bool, lightColor color.Color) (color.Color, color.Color) {
	attributes := c.videoAttribute
	if topBit {
		attributes >>= 4
	}

	inverse := attributes&videxUltratermAttributesInverse != 0
	highlight := attributes&videxUltratermAttributesHighlight != 0

	var clearColor color.Color = color.Black
	setColor := lightColor

	if !highlight {
		r, g, b, a := setColor.RGBA()
		setColor = color.NRGBA64{
			uint16(r>>1 + r>>2),
			uint16(g>>1 + g>>2),
			uint16(b>>1 + b>>2),
			uint16(a)}
	}

	if inverse {
		temp := setColor
		setColor = clearColor
		clearColor = temp
	}

	return clearColor, setColor

}

func (c *CardVidexUltraterm) buildImage(light color.Color) *image.RGBA {
	params := c.mc6845.ImageData()
	width, height := params.DisplayedWidthHeight(videxUltratermCharWidth)
	if (width == 0) || (height == 0) {
		// No image available
		size := image.Rect(0, 0, 3, 3)
		img := image.NewRGBA(size)
		img.Set(1, 1, color.White)
		return img
	}
	ms := time.Now().Nanosecond() / (1000 * 1000) // Host time, used for the cursor blink

	size := image.Rect(0, 0, width, height)
	img := image.NewRGBA(size)

	upperClearColor, upperSetColor := c.colorsPerAttributes(true, light)
	lowerClearColor, lowerSetColor := c.colorsPerAttributes(false, light)
	altChar := c.videoAttribute&videxUltratermAttributesAlternateChar != 0

	sramMask := videxUltratermSram256Mask
	is512mode := c.modeControl&videxUltratermMCPSRamAdressFormat == 0
	if is512mode {
		sramMask = videxUltratermSram512Mask
	}

	params.IterateScreen(func(address uint16, charLine uint8,
		cursorMode uint8, displayEnable bool,
		column uint8, y int) {

		if !displayEnable {
			return
		}

		bits := uint8(0)
		colorOn := lowerSetColor
		colorOff := lowerClearColor
		char := c.sram[address&sramMask]
		if char&0x80 != 0 {
			colorOn = upperSetColor
			colorOff = upperClearColor
		}

		romIndex := (uint16(char&0x7f) << 4) + uint16(charLine)
		if !altChar {
			romIndex += 2048
		}
		bits = c.charGen[romIndex]

		// Cursor
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

		x := int(column) * int(videxUltratermCharWidth)
		color := colorOff
		for i := 0; i < int(videxUltratermCharWidth-1); i++ {
			pixel := (bits & 0x80) != 0
			if pixel {
				color = colorOn
			} else {
				color = colorOff
			}
			img.Set(x, y, color)
			bits <<= 1
			x++
		}

		// The ninth bit: blank or repetition of the last bit
		if char&0x7f < 0x20 {
			img.Set(x, y, color)
		} else {
			img.Set(x, y, colorOff)
		}
	})

	return img
}

func (c *CardVidexUltraterm) getText() string {
	text := ""
	params := c.mc6845.ImageData()

	sramMask := videxUltratermSram256Mask
	is512mode := c.modeControl&videxUltratermMCPSRamAdressFormat == 0
	if is512mode {
		sramMask = videxUltratermSram512Mask
	}

	address := params.FirstChar
	for line := uint8(0); line < params.Lines; line++ {
		for column := uint8(0); column < params.Columns; column++ {
			char := c.sram[address&sramMask]
			text += string(char)
			address++
		}
		text = strings.TrimRight(text, " ")
		text += "\n"
	}
	return text
}

//lint:ignore U1000 Ignore function used for debugging
func (c *CardVidexUltraterm) dumpState() {
	data := c.mc6845.ImageData()
	width, height := data.DisplayedWidthHeight(videxUltratermCharWidth)
	is512mode := c.modeControl&videxUltratermMCPSRamAdressFormat == 0
	sramPage256 := c.modeControl & videxUltratermMCPSramPageMask
	mhz := "17.430"
	if c.modeControl&videxUltratermMCPClockFrequency != 0 {
		mhz = "28.7595"
	}

	flags := c.a.mmu.Peek(0x7f8 + uint16(c.slot))

	fmt.Printf("%vx%v %vx%v %vx%v  512:%v,%v 256page:%v %v MHz +-%v Flags: %v\n",
		videxUltratermCharWidth, data.CharLines,
		data.Columns, data.Lines, width, height, is512mode, c.sramPage512, sramPage256, mhz,
		data.AdjustLines, flags)
}
