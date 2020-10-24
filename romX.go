package izapple2

import (
	"fmt"

	"github.com/ivanizag/izapple2/core6502"
)

/*
RomX from https://theromexchange.com/
This complement uses the RomX API spec to switch main ROM and character generator ROM

See:
	https://theromexchange.com/documentation/ROM%20X%20API%20Reference.pdf
	https://theromexchange.com/downloads/ROM%20X%2020-10-22.zip

It is not enough to intercept the ROM accesses. RomX intercept the 74LS138 in
position F12, that has access to the full 0xc000-0xf000 on the Apple II+
*/

type romX struct {
	a              *Apple2
	memory         core6502.Memory
	activationStep int
	systemBank     uint8
	textBank       uint8
}

var romXActivationSequence = []uint16{0xcaca, 0xcaca, 0xcafe}

const (
	setupBank                    = uint8(0)
	romXSetSystemBankBaseAddress = uint16(0xcef0)
	romXSetTextBankBaseAddress   = uint16(0xcfd0)
	romXDefaultSystemBankAddress = uint16(0xd034)
	romXDefaultTextBankAddress   = uint16(0xd02e)

	// Unknown
	romXFirmwareMark0Address = uint16(0xdffe)
	romXFirmwareMark0Value   = uint8(0x4a)
	romXFirmwareMark1Address = uint16(0xdfff)
	romXFirmwareMark1Value   = uint8(0xcd)
)

func newRomX(a *Apple2, memory core6502.Memory) *romX {
	var rx romX
	rx.a = a
	rx.memory = memory
	rx.systemBank = 0
	rx.textBank = 0
	return &rx
}

func (rx *romX) Peek(address uint16) uint8 {
	intercepted, value := rx.interceptAccess(address)
	if intercepted {
		return value
	}
	return rx.memory.Peek(address)
}

func (rx *romX) PeekCode(address uint16) uint8 {
	//intercepted, value := rx.interceptAccess(address)
	//if intercepted {
	//	return value
	//}
	return rx.memory.PeekCode(address)
}

func (rx *romX) Poke(address uint16, value uint8) {
	rx.interceptAccess(address)
	rx.memory.Poke(address, value)
}

func (rx *romX) interceptAccess(address uint16) (bool, uint8) {
	// Intercept only $C080 to $FFFF as seen by the F12 chip
	if address < 0xc080 {
		return false, 0
	}

	// Setup mode when the setup bank is active
	if rx.systemBank == setupBank {
		switch address {
		case romXDefaultSystemBankAddress:
			fmt.Printf("[romX]Peek in $%04x, current system bank %v\n", address, rx.systemBank)
			return true, 0xe0 + rx.systemBank
		case romXDefaultTextBankAddress:
			fmt.Printf("[romX]PeeK in $%04x, current text bank %v\n", address, rx.textBank)
			return true, 0xd0 + rx.textBank
		case romXFirmwareMark0Address:
			fmt.Printf("[romX]Peek in $%04x, ???\n", address)
			return true, romXFirmwareMark0Value
		case romXFirmwareMark1Address:
			fmt.Printf("[romX]Peek in $%04x, ???\n", address)
			return true, romXFirmwareMark1Value
		}

		if address&0xfff0 == romXSetSystemBankBaseAddress {
			rx.systemBank = uint8(address & 0xf)
			fmt.Printf("[romX]System bank set to %v\n", rx.systemBank)
		} else if address&0xfff0 == romXSetTextBankBaseAddress {
			rx.textBank = uint8(address & 0xf)
			fmt.Printf("[romX]Text bank set to %v\n", rx.textBank)
		} else if address < 0xe000 {
			fmt.Printf("[romX]Peek in $%04x\n", address)
		}

		return false, 0
	}

	// Activation sequence detection
	if address == romXActivationSequence[rx.activationStep] {
		rx.activationStep++
		//fmt.Printf("[romX]Activation step %v\n", rx.activationStep)
		if rx.activationStep == len(romXActivationSequence) {
			// Activation sequence completed
			rx.systemBank = setupBank
			rx.activationStep = 0
			//			rx.a.cpu.SetTrace(true)
			fmt.Printf("[romX]System bank set to 0, %v\n", rx.systemBank)
		}
	} else {
		rx.activationStep = 0
	}

	return false, 0
}
