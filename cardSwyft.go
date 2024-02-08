package izapple2

/*
	Swyft card for Apple IIe

See:
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Other/IAI%20SwyftCard/

*/

/*
"SwyftCard Hardware  Theory of Operation". SwyftCard manual, page 98:

The SwyftCard is a plug-in card for the Apple /Ie that operates in
slot 3. The card contains three integrated circuits which provide a
power-on reset circuit, storage for the SwyftCard program, and control
signals for the card. The card operates by asserting the Apple IIe bus
signal INH' which disables the built-in ROM and enables the SwyftCard
ROM. This permits the SwyftCard program to take over the system at
power-on and run the SwyftCard program. (Please refer to the
schematic.)

The LM311 voltage comparator is connected to provide the power-on
reset function. When the Apple lIe is first turned on, the power-on
reset circuit resets the PAL, turning on the SwyftCard and disabling
the Apple IIe internal ROM. The power-on reset circuit must be
provided because the existing Apple IIe reset function is used by
many Apple lie programs for a "warm start": if Apple lie reset always
started the SwyftCard, other programs could not use the "warm start."

The 27128 PROM is used to store the SwyftCard program. The PROM
contains 16384 bytes which are mapped into the address space
$DOOO - $FFFF. Since the address space is only 12 Kbytes, there are
two 4 Kbyte sections of the PROM mapped into the address space
$DOOO-$DFFF.

The card is controlled by the PAL. When the SwyftCard is active, the
PAL asserts the INH' signal, enables the PROM, and bank switches
the $DOOO-$DFFF address space. The card is controlled by two soft
switches. The soft switches are controlled by accessing the following
memory locations with either a read or a write operation.

	$COBO - SwyftCard active, Bank 1
	$COB1 - SwyftCard inactive, Bank 1
	$COB2 - SwyftCard active, Bank 2

When the power-on reset circuit asserts the RES signal on Pin 3 of the
PAL, the SwyftCard is made active in Bank 1. Accessing location
$COB1 deactivates the SwyftCard for normal Apple IIe operation.

The INH' line is driven by a tri-state driver, so if another card in the
Apple /Ie asserts the IINH' signal there will not be a bus contention.
However, there will be a bus contention on the data bus if another card
attempts to control the bus while the SwyftCard is active.

The Cx00 rom is not used. The card is expected to be installed in
slot 3 of an Apple IIe with the 80 column firmware already present.

*/

// CardSwyft represents a Swyft card
type CardSwyft struct {
	cardBase
	bank2 bool
	rom   []uint8
}

func newCardSwyftBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "SwyftCard",
		description: "Card with the ROM needed to run the Swyftcard word processing system",
		requiresIIe: true,
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardSwyft

			// Load main ROM replacement
			data, _, err := LoadResource("<internal>/SwyftCard ROM.bin")
			if err != nil {
				return nil, err
			}
			c.rom = data
			return &c, nil
		},
	}
}

func (c *CardSwyft) assign(a *Apple2, slot int) {
	if slot != 3 {
		panic("SwyftCard must be installed in slot 3")
	}

	c.addCardSoftSwitchRW(0, func() uint8 {
		a.mmu.inhibitROM(c)
		c.bank2 = false
		return 0x55
	}, "SWYFTONBANK1")

	c.addCardSoftSwitchRW(1, func() uint8 {
		a.mmu.inhibitROM(nil)
		c.bank2 = false
		return 0x55
	}, "SWYFTOFFBANK1")

	c.addCardSoftSwitchRW(2, func() uint8 {
		a.mmu.inhibitROM(c)
		c.bank2 = true
		return 0x55
	}, "SWYFTONBANK2")

	c.cardBase.assign(a, slot)
	a.mmu.inhibitROM(c)
}

func (c *CardSwyft) translateAddress(address uint16) uint16 {
	/*
		The four 4k sections of the 16k ROM image are mapped:
			D000-DFFF (page 1)
			D000-DFFF (page 2)
			E000-EFFF
			F000-FFFF
	*/
	if address >= 0xE000 {
		return address - 0xE000 + 0x2000
	}
	if !c.bank2 {
		return address - 0xD000
	}
	return address - 0xD000 + 0x1000
}

func (c *CardSwyft) peek(address uint16) uint8 {
	return c.rom[c.translateAddress(address)]
}

func (c *CardSwyft) poke(address uint16, value uint8) {
	// Nothing
}

func (c *CardSwyft) setBase(base uint16) {
	// Nothing
}
