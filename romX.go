package izapple2

import (
	"fmt"

	"github.com/ivanizag/iz6502"
)

/*
RomX from https://theromexchange.com/
This complement uses the RomX API spec to switch main ROM and character generator ROM

Only the font switch is implemented

See:
	https://theromexchange.com/documentation/ROM%20X%20API%20Reference.pdf
	https://theromexchange.com/downloads/ROM%20X%2020-10-22.zip
	https://theromexchange.com/documentation/romxce/ROMXce%20API%20Reference.pdf

For romX:
It is not enough to intercept the ROM accesses. RomX intercept the 74LS138 in
position F12, that has access to the full 0xc000-0xf000 on the Apple II+

Firmware:
	- It first copies $D000-$DFFF to $6000 and runs there.

go run *.go -rom ROMX.FIRM.dump -disk ROM\ X\ 20-10-22.dsk


*/

type romX struct {
	a              *Apple2
	memory         iz6502.Memory
	activationStep int
	systemBank     uint8
	mainBank       uint8
	tempBank       uint8
	textBank       uint8
	debug          bool
}

var romXActivationSequence = []uint16{0xcaca, 0xcaca, 0xcafe}
var romXceActivationSequence = []uint16{0xfaca, 0xfaca, 0xfafe}

const (
	romxSetupBank                    = uint8(0)
	romXPlusSetSystemBankBaseAddress = uint16(0xcef0)
	romXPlusSetTextBankBaseAddress   = uint16(0xcfd0)

	// Unknown
	// romXFirmwareMark0Address = uint16(0xdffe)
	// romXFirmwareMark0Value   = uint8(0x4a)
	// romXFirmwareMark1Address = uint16(0xdfff)
	// romXFirmwareMark1Value   = uint8(0xcd)

	romXceSelectTempBank  = uint16(0xf850)
	romXceSelectMainBank  = uint16(0xf851)
	romXceSetTempBank     = uint16(0xf830) // 16 positions
	romXceSetMainBank     = uint16(0xf800) // 16 positions
	romXcePresetTextBank  = uint16(0xf810) // 16 positions
	romXceMCP7940SDC      = uint16(0xf860) // 16 positions
	romXceLowerUpperBanks = uint16(0xf820) // 16 positions

	romXGetDefaultSystemBank = uint16(0xd034) // $00 to $0f
	romXGetDefaultTextBank   = uint16(0xd02e) // $10 to $1f
	romXGetCurrentBootDelay  = uint16(0xdeca) // $00 to $0f

	/*
		romXceEntryPointSetClock      = uint16(0xc803)
		romXceEntryPointReadClock     = uint16(0xc803)
		romXceEntryPointLauncherToRam = uint16(0xdfd9)
		romXceEntryPointLauncher      = uint16(0xdfd0)
	*/
)

func newRomX(a *Apple2) (*romX, error) {
	var rx romX
	rx.a = a
	rx.memory = a.mmu
	rx.systemBank = 1
	rx.mainBank = 1
	rx.tempBank = 1
	rx.textBank = 0
	rx.debug = true

	if a.isApple2e {
		err := a.cg.load("<internal>/ROMXce Production 1Mb Text ROM V5.bin")
		if err != nil {
			return nil, err
		}
	}

	// Intercept all memory accesses
	a.cpu.SetMemory(&rx)
	return &rx, nil
}

func (rx *romX) Peek(address uint16) uint8 {
	intercepted, value := rx.interceptAccess(address)
	if intercepted {
		return value
	}
	return rx.memory.Peek(address)
}

func (rx *romX) PeekCode(address uint16) uint8 {
	intercepted, value := rx.interceptAccess(address)
	if intercepted {
		return value
	}
	return rx.memory.PeekCode(address)
}

func (rx *romX) Poke(address uint16, value uint8) {
	rx.interceptAccess(address)
	rx.memory.Poke(address, value)
}

func (rx *romX) logf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("[romX]%s\n", msg)
}

func (rx *romX) interceptAccess(address uint16) (bool, uint8) {
	// Intercept only $C080 to $FFFF as seen by the F12 chip
	if address < 0xc080 {
		return false, 0
	}

	// Setup mode when the setup bank is active
	if rx.systemBank == romxSetupBank {

		// Range commands
		nibble := uint8(address & 0xf)
		switch address & 0xfff0 {
		case romXceSetMainBank:
			rx.mainBank = nibble
			rx.logf("Main bank set to $%x", nibble)
		case romXcePresetTextBank:
			textBank := int(nibble)
			rx.a.cg.setPage(textBank)
			rx.logf("[romX]Text bank set to $%x", nibble)
		case romXceLowerUpperBanks:
			rx.logf("Configure lower upper banks $%x", address)
		case romXceSetTempBank:
			rx.tempBank = nibble
			rx.logf("Temp bank set to $%x", nibble)
		case romXceMCP7940SDC:
			rx.logf("Configure MCP7940 $%x", address)
		}

		// More commands
		switch address {
		case romXceSelectTempBank:
			rx.systemBank = rx.tempBank
			rx.logf("System bank set to temp bank $%x", rx.systemBank)
		case romXceSelectMainBank:
			rx.systemBank = rx.mainBank
			rx.logf("System bank set to main bank $%x", rx.systemBank)
		}

		// Queries
		switch address {
		case romXGetDefaultSystemBank:
			bank := rx.systemBank
			rx.logf("Peek in $%04x, current system bank %v", address, bank)
			return true, bank
		case romXGetDefaultTextBank:
			page := uint8(rx.a.cg.getPage() & 0xf)
			rx.logf("PeeK in $%04x, current text bank %v", address, page)
			return true, 0x10 + page
		case romXGetCurrentBootDelay:
			delay := uint8(5) // We don't care
			rx.logf("PeeK in $%04x, current boot delay %v", address, delay)
			return true, delay
		}

		return false, 0
	}

	// Activation sequence detection
	if address == romXceActivationSequence[rx.activationStep] {
		rx.activationStep++
		rx.logf("Activation step %v", rx.activationStep)
		if rx.activationStep == len(romXActivationSequence) {
			// Activation sequence completed
			rx.systemBank = romxSetupBank
			rx.activationStep = 0
			rx.logf("System bank set to 0, %v", rx.systemBank)
		}
	} else {
		rx.activationStep = 0
	}

	return false, 0
}

func setupRomX(a *Apple2) error {
	_, err := newRomX(a)
	if err != nil {
		return err
	}
	return nil
}
