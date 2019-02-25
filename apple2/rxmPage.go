package apple2

type rxmPage struct {
	data     [256]uint8
	observer func(address uint8, isWrite bool)
}

type ramPage struct {
	rxmPage
}

type romPage struct {
	rxmPage
}

func (p *rxmPage) Peek(address uint8) uint8 {
	p.touch(address, false)
	return p.data[address]
}

func (p *rxmPage) internalPeek(address uint8) uint8 {
	return p.data[address]
}

func (p *rxmPage) Poke(address uint8, value uint8) {
	p.touch(address, true)
	p.data[address] = value
}

func (p *rxmPage) touch(address uint8, isWrite bool) {
	if p.observer != nil {
		p.observer(address, isWrite)
	}
}

func (p *romPage) Poke(address uint8, value uint8) {
	// Do nothing
}

func (p *romPage) burn(address uint8, value uint8) {
	p.data[address] = value
}
