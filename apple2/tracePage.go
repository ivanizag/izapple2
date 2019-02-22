package apple2

import "fmt"

type tracePage struct {
	page uint8
}

func (p *tracePage) Peek(address uint8) uint8 {
	fmt.Printf("Read on address 0x%02x%02x\n", p.page, address)
	panic(address)
	return 0xcc
}

func (p *tracePage) Poke(address uint8, value uint8) {
	fmt.Printf("Write on address 0x%02x%02x\n", p.page, address)
	panic(address)
}
