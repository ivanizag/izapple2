package apple2

type textPages struct {
	pages [4]textPage
}

type textPage struct {
	dirty bool
	data  [256]uint8
}

func (p *textPage) Peek(address uint8) uint8 {
	return p.data[address]
}

func (p *textPage) Poke(address uint8, value uint8) {
	p.data[address] = value
	// Note: we could avoid setting dirty on the 16 blocks of 8 hidden bytes
	p.dirty = true
}

func (tp *textPages) read(column uint8, line uint8) uint8 {
	page, address := tp.charAddress(column, line)
	return tp.pages[page].Peek(address)
}

func (tp *textPages) write(column uint8, line uint8, value uint8) {
	page, address := tp.charAddress(column, line)
	tp.pages[page].Poke(address, value)
}

func (tp *textPages) charAddress(column uint8, line uint8) (page uint8, address uint8) {
	page = (line % 8) / 2
	address = column + (line/8)*40 + (line%2)*128
	return
}

func (tp *textPages) strobe() bool {
	// Thread safe. May just mark more dirties than needed.
	dirty := false
	for i := 0; i < 4; i++ {
		if tp.pages[i].dirty {
			dirty = true
			tp.pages[i].dirty = false
		}
	}
	return dirty
}
