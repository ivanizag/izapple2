package izapple2

import (
	"strings"
	"testing"
)

func Test2cBootsDOS33(t *testing.T) {
	testBoots(t, "2c", "<internal>/dos33.dsk", nil, 100_000_000, "DOS VERSION 3.3", "\n]", testTextMode40)
}

func Test2cBootsProDOS(t *testing.T) {
	testBoots(t, "2c", "<internal>/ProDOS_2_4_3.po", nil, 100_000_000, "BITSY  BYE", "NEW VOL", testTextMode40)
}

func Test2cRefusesCards(t *testing.T) {
	for name, overrides := range map[string]map[string]string{
		"a card on a slot":             {"s4": "mockingboard"},
		"another card on slot 6":       {"s6": "smartport"},
		"no language card":             {"s0": "empty"},
		"the aux slot of the IIe":      {"ramworks": "8192"},
		"a card on the slot of the LC": {"s0": "saturn"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := CreateAppleFromModel("2c", overrides, nil)
			if err == nil {
				t.Error("expected the configuration to be refused")
			}
		})
	}
}

func Test2cRefusesPositionalBlockDevices(t *testing.T) {
	// They would go to a smartport card on slot 7
	_, err := CreateAppleFromModel("2c", nil, []string{"<internal>/ProDOS_2_4_3.po", "<internal>/A2DeskTop-1.4-en_800k.2mg"})
	if err == nil || !strings.Contains(err.Error(), "no slots") {
		t.Errorf("expected the IIc to refuse the block devices, got %v", err)
	}
}

/*
runInterruptProgram runs a program on a IIc once its firmware is ready. The
handler is at $0380 and counts in $0350 the interrupts it gets. A JSR $0340
calls the firmware entry point stored in $0360.
*/
func runInterruptProgram(t *testing.T, program []uint8, handler []uint8, frames uint64) *Apple2 {
	t.Helper()
	a, err := CreateAppleFromModel("2c", map[string]string{"s6": "empty"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	a.Init()
	a.RunCycles(2_000_000)

	a.mmu.pokeRange(0x0300, program)
	a.mmu.pokeRange(0x0340, []uint8{0x6c, 0x60, 0x03}) // JMP ($0360)
	a.mmu.pokeRange(0x0380, handler)
	a.cpu.SetPC(0x0300)
	a.RunCycles(frames * screenDrawCycles)
	return a
}

// vblHardwareProgram takes the IRQ vector on the language card RAM, away
// from the firmware, and turns on the VBL interrupts if asked to
func vblHardwareProgram(enable bool) []uint8 {
	enableVBL := []uint8{0xea, 0xea, 0xea} // NOP
	if enable {
		enableVBL = []uint8{0x8d, 0x5b, 0xc0} // STA $C05B, ENVBL
	}
	program := []uint8{
		0x78,             // SEI
		0xad, 0x8b, 0xc0, // LDA $C08B, LC RAM read and write
		0xad, 0x8b, 0xc0, // LDA $C08B
		0xa9, 0x80, // LDA #$80
		0x8d, 0xfe, 0xff, // STA $FFFE
		0xa9, 0x03, // LDA #$03
		0x8d, 0xff, 0xff, // STA $FFFF
		0x9c, 0x50, 0x03, // STZ $0350
		0x8d, 0x79, 0xc0, // STA $C079, IOUDIS off
	}
	program = append(program, enableVBL...)
	return append(program,
		0x8d, 0x78, 0xc0, // STA $C078, IOUDIS on
		0x58,       // CLI
		0x80, 0xfe, // BRA *
	)
}

var vblHardwareHandler = []uint8{
	0xee, 0x50, 0x03, // INC $0350
	0xad, 0x19, 0xc0, // LDA $C019, the VBL interrupt flag
	0x8d, 0x51, 0x03, // STA $0351
	0xad, 0x70, 0xc0, // LDA $C070, resets the VBL interrupt
	0xad, 0x19, 0xc0, // LDA $C019
	0x8d, 0x52, 0x03, // STA $0352
	0x40, // RTI
}

func Test2cVBLInterrupts(t *testing.T) {
	a := runInterruptProgram(t, vblHardwareProgram(true), vblHardwareHandler, 10)
	if count := a.mmu.Peek(0x0350); count < 9 || count > 11 {
		t.Errorf("expected an interrupt per frame, 10, got %v", count)
	}
	if flag := a.mmu.Peek(0x0351); flag != ssOn {
		t.Errorf("expected $C019 to show the VBL interrupt, got $%02x", flag)
	}
	if flag := a.mmu.Peek(0x0352); flag != ssOff {
		t.Errorf("expected $C070 to reset the VBL interrupt, got $%02x", flag)
	}
}

func Test2cNoVBLInterruptsWhenDisabled(t *testing.T) {
	a := runInterruptProgram(t, vblHardwareProgram(false), vblHardwareHandler, 10)
	if count := a.mmu.Peek(0x0350); count != 0 {
		t.Errorf("expected no interrupts with VBL disabled, got %v", count)
	}
}

func Test2cVBLInterruptsFromTheMouseFirmware(t *testing.T) {
	// The firmware owns the VBL interrupt, SetMouse turns it on and the
	// firmware passes it to the handler on $03FE
	program := []uint8{
		0x78,       // SEI
		0xa9, 0x80, // LDA #$80
		0x8d, 0xfe, 0x03, // STA $03FE
		0xa9, 0x03, // LDA #$03
		0x8d, 0xff, 0x03, // STA $03FF
		0x9c, 0x50, 0x03, // STZ $0350
		0xad, 0x12, 0xc4, // LDA $C412, the SetMouse entry
		0x8d, 0x60, 0x03, // STA $0360
		0xa9, 0xc4, // LDA #$C4
		0x8d, 0x61, 0x03, // STA $0361
		0xa9, 0x08, // LDA #$08, VBL interrupts with the mouse off
		0xa2, 0xc4, // LDX #$C4
		0xa0, 0x40, // LDY #$40
		0x20, 0x40, 0x03, // JSR $0340
		0x58,       // CLI
		0x80, 0xfe, // BRA *
	}
	handler := []uint8{
		0xee, 0x50, 0x03, // INC $0350
		0x40, // RTI
	}
	a := runInterruptProgram(t, program, handler, 10)
	if count := a.mmu.Peek(0x0350); count < 9 || count > 11 {
		t.Errorf("expected an interrupt per frame, 10, got %v", count)
	}
}

func Test2c80Columns(t *testing.T) {
	a, err := CreateAppleFromModel("2c", map[string]string{"s6": "empty"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := &strobedKeys{}
	a.SetKeyboardProvider(keys)
	a.Init()
	a.RunCycles(2_000_000)

	// Without a disk the firmware waits for one, Ctrl-Reset goes to BASIC
	a.reset()
	a.RunCycles(1_000_000)
	typeLines(t, a, keys, "PR#3", `PRINT "ON 80 COLUMNS"`)
	a.RunCycles(1_000_000)

	if !a.io.isSoftSwitchActive(ioFlag80Col) {
		t.Error("expected the 80 column mode to be on")
	}
	if !strings.Contains(a.ScreenText(), "ON 80 COLUMNS") {
		t.Errorf("expected the text on the screen, got:\n%v", a.ScreenText())
	}
}

// testMouse is a host pointer that the tests move
type testMouse struct {
	x, y    uint16
	pressed bool
}

func (m *testMouse) ReadMouse() (uint16, uint16, bool) {
	return m.x, m.y, m.pressed
}

/*
mouseFirmwareProgram calls InitMouse, turns the mouse on with SetMouse and
calls ReadMouse in a loop, for the position to be on the screen holes.
*/
var mouseFirmwareProgram = []uint8{
	0x78,       // SEI
	0xa9, 0xc4, // LDA #$C4
	0x8d, 0x61, 0x03, // STA $0361
	0xad, 0x19, 0xc4, // LDA $C419, InitMouse
	0x8d, 0x60, 0x03, // STA $0360
	0xa2, 0xc4, // LDX #$C4
	0xa0, 0x40, // LDY #$40
	0x20, 0x40, 0x03, // JSR $0340
	0xad, 0x12, 0xc4, // LDA $C412, SetMouse
	0x8d, 0x60, 0x03, // STA $0360
	0xa9, 0x01, // LDA #$01, the mouse on, no interrupts for the program
	0xa2, 0xc4, // LDX #$C4
	0xa0, 0x40, // LDY #$40
	0x20, 0x40, 0x03, // JSR $0340
	0xad, 0x14, 0xc4, // LDA $C414, ReadMouse
	0x8d, 0x60, 0x03, // STA $0360
	0xa2, 0xc4, // LDX #$C4
	0xa0, 0x40, // LDY #$40
	0x20, 0x40, 0x03, // JSR $0340
	0x80, 0xf1, // BRA to ReadMouse
}

// mousePosition is the position ReadMouse left on the screen holes
func mousePosition(a *Apple2) (int, int) {
	x := int(a.mmu.Peek(0x047c)) + int(a.mmu.Peek(0x057c))<<8
	y := int(a.mmu.Peek(0x04fc)) + int(a.mmu.Peek(0x05fc))<<8
	return x, y
}

func Test2cMouseMoves(t *testing.T) {
	a, err := CreateAppleFromModel("2c", map[string]string{"s6": "empty"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mouse := &testMouse{}
	a.SetMouseProvider(mouse)
	a.Init()
	a.RunCycles(2_000_000)
	a.mmu.pokeRange(0x0300, mouseFirmwareProgram)
	a.mmu.pokeRange(0x0340, []uint8{0x6c, 0x60, 0x03}) // JMP ($0360)
	a.cpu.SetPC(0x0300)
	a.RunCycles(5 * screenDrawCycles)

	// The middle of the window, half of the screen of steps
	mouse.x, mouse.y = 32768, 32768
	a.RunCycles(60 * screenDrawCycles)
	if x, y := mousePosition(a); x != mouseStepsX/2 || y != mouseStepsY/2 {
		t.Errorf("expected the mouse on %v,%v, got %v,%v", mouseStepsX/2, mouseStepsY/2, x, y)
	}

	// Back to the corner
	mouse.x, mouse.y = 0, 0
	a.RunCycles(60 * screenDrawCycles)
	if x, y := mousePosition(a); x != 0 || y != 0 {
		t.Errorf("expected the mouse back on 0,0, got %v,%v", x, y)
	}
}

func Test2cMouseMeetsOnTheEdges(t *testing.T) {
	a, err := CreateAppleFromModel("2c", map[string]string{"s6": "empty"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// The moves before the mouse is on are lost, the pointers start apart
	mouse := &testMouse{x: 65535, y: 65535}
	a.SetMouseProvider(mouse)
	a.Init()
	a.RunCycles(2_000_000)
	a.mmu.pokeRange(0x0300, mouseFirmwareProgram)
	a.mmu.pokeRange(0x0340, []uint8{0x6c, 0x60, 0x03}) // JMP ($0360)
	a.cpu.SetPC(0x0300)
	a.RunCycles(60 * screenDrawCycles)

	// Pushing on the edge moves the mouse of the IIc
	if x, y := mousePosition(a); x <= mouseStepsX/2 || y <= mouseStepsY/2 {
		t.Errorf("expected the mouse pushed towards the edge, got %v,%v", x, y)
	}

	// Beyond the screen, up to the clamp of the firmware, and back to 0,0
	mouse.x, mouse.y = 0, 0
	a.RunCycles(120 * screenDrawCycles)
	if x, y := mousePosition(a); x != 0 || y != 0 {
		t.Errorf("expected the mouse back on 0,0, got %v,%v", x, y)
	}
}
