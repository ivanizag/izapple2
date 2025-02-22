package component

/*
	MC6845 CRT Controller
	See:
		Motorola MC6845 datasheet

	Pins:
		RW, RS, D0-D7: Read() and Write()
		MA0-13, RA04, CURSOR, DE: MC6845RasterCallBack()
*/

type MC6845 struct {
	reg [18]uint8 // Internal registers R0 to R17
	sel uint8     // Selected address register AR
}

func (m *MC6845) Read(rs bool) uint8 {
	if !rs {
		// AR is not readable
		return 0x00
	} else if m.sel >= 14 && m.sel <= 17 {
		// Only R14 to R17 are readable
		// Should we mask R14 and R16?
		return m.reg[m.sel]
	}
	return 0x00
}

func (m *MC6845) Write(rs bool, value uint8) {
	if !rs {
		// AR is 5 bits
		// What happens if AR > 17 ?
		m.sel = value & 0x1f
	} else if m.sel <= 15 {
		// R0 to R15 are writable

		if m.sel == 1 && value == 144 {
			// Horrible hack for the mode 6.
			m.reg[m.sel] = 160
		} else {
			m.reg[m.sel] = value
		}
	}
}

func (m *MC6845) ImageData() MC6845ImageData {
	var data MC6845ImageData

	data.FirstChar = uint16(m.reg[12]&0x3f)<<8 + uint16(m.reg[13])
	data.CharLines = (m.reg[9] + 1) & 0x1f
	data.Columns = m.reg[1]
	data.Lines = m.reg[6] & 0x7f
	data.AdjustLines = m.reg[5] & 0x1f

	data.cursorPos = uint16(m.reg[14]&0x3f)<<8 + uint16(m.reg[15])
	data.cursorStart = m.reg[10] & 0x1f
	data.cursorEnd = m.reg[11] & 0x1f
	data.cursorMode = (m.reg[10] >> 5) & 0x03 // Bits 6 and 5

	data.InterlaceMode = m.reg[8] & 0x03 // Bits 1 and 0
	return data
}

const (
	MC6845CursorFixed = uint8(0)
	MC6845CursorNone  = uint8(1)
	MC6845CursorFast  = uint8(2)
	MC6845CursorSlow  = uint8(3)
)

type MC6845ImageData struct {
	FirstChar   uint16 // 14 bits, address of the firt char on the first line
	CharLines   uint8  // 5 bits, lines par character
	Columns     uint8  // 8 bits, chars per line
	Lines       uint8  // 7 bits, char lines per screen
	AdjustLines uint8  // 5 bits, extra blank lines

	cursorPos   uint16 // 14 bits, address? of the cursor position
	cursorStart uint8  // 5 bits, cursor starting char row
	cursorEnd   uint8  // 5 bits, cursor ending char row
	cursorMode  uint8  // 2 bits, cursor mode

	InterlaceMode uint8 // 2 bit, interlace mode

}

func (data *MC6845ImageData) DisplayedWidthHeight(charWidth uint8) (int, int) {
	return int(data.Columns) * int(charWidth),
		int(data.Lines)*int(data.CharLines) + int(data.AdjustLines)
}

type MC6845RasterCallBack func(address uint16, charLine uint8, // Lookup in char ROM
	cursorMode uint8, displayEnable bool, // Modifiers
	column uint8, y int) // Position in screen

func (data *MC6845ImageData) IterateScreen(callBack MC6845RasterCallBack) {
	lineAddress := data.FirstChar
	y := 0
	var address uint16
	for line := uint8(0); line < data.Lines; line++ {
		for charLine := uint8(0); charLine < data.CharLines; charLine++ {
			address = lineAddress // Back to the first char of the line
			for column := uint8(0); column < data.Columns; column++ {
				cursorMode := MC6845CursorNone
				isCursor := (address == data.cursorPos) &&
					(charLine >= data.cursorStart) &&
					(charLine <= data.cursorEnd)
				if isCursor {
					cursorMode = data.cursorMode
				}

				callBack(address, charLine, cursorMode, true, column, y)
				address = (address + 1) & 0x3fff // 14 bits
			}
			y++
		}
		lineAddress = address
	}
	for adjust := uint8(0); adjust <= data.AdjustLines; adjust++ {
		for column := uint8(0); column < data.Columns; column++ {
			callBack(0, 0, MC6845CursorNone, false, column, y) // lines with display not enabled
		}
		y++
	}
}
