package apple2

type ramPage struct {
	data [256]uint8
}

type romPage struct {
	data [256]uint8
}

func (p *ramPage) Peek(address uint8) uint8 {
	return p.data[address]
}

func (p *ramPage) Poke(address uint8, value uint8) {
	p.data[address] = value
}

func (p *romPage) Peek(address uint8) uint8 {
	return p.data[address]
}

func (p *romPage) Poke(address uint8, value uint8) {
	// Do nothing
}

func (p *romPage) burn(address uint8, value uint8) {
	p.data[address] = value
}
