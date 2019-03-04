package apple2

type unassignedPage struct {
	page uint8
}

func (p *unassignedPage) Peek(address uint8) uint8 {
	//fmt.Printf("Read on address 0x%02x%02x\n", p.page, address)
	//panic(address)
	return 0xdd
}

func (p *unassignedPage) Poke(address uint8, value uint8) {
	//fmt.Printf("Write on address 0x%02x%02x\n", p.page, address)
	//panic(address)
}
