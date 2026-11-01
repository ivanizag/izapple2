package izapple2

import (
	"fmt"
)

/*
Apple IIc. It is an Apple IIe enhanced with the 80 column card and its 64Kb,
the language card, the disk controller, two serial ports and the mouse built
in. The firmware of all of them is in the ROM: $C100 to $CFFF is always the
internal ROM and there are no slots.

The IOU of the IIc handles the mouse and the vertical blank, both able to
interrupt the CPU. Its soft switches take $C058 to $C05F, the annunciators
on the IIe, while IOUDIS is off.

See:
	Apple IIc Technical Reference Manual, chapter 9 and appendix E.
		https://archive.org/details/a2ctrm
	MAME, apple2e.cpp
		https://github.com/mamedev/mame/blob/master/src/mame/apple/apple2e.cpp
*/

// irqSourceIOU is the source of the IRQ requests of the IOU, after the slots
const irqSourceIOU = 8

// iouApple2c is the state of the IOU of the IIc not shared with the IIe
type iouApple2c struct {
	a *Apple2

	/*
		iouSwitchesDisabled is IOUDIS, IOU disable. The IOU keeps working and
		interrupting, only its switches on $C058 to $C05F are disabled: the
		annunciators and the double hires switch take their place, as on the
		IIe. It is on after power on.
	*/
	iouSwitchesDisabled bool

	vblMask   bool // VBL interrupts enabled
	xyMask    bool // Mouse movement interrupts enabled
	x0Falling bool // The X0 interrupt is on the falling edge, on the rising if false
	y0Falling bool // The Y0 interrupt is on the falling edge, on the rising if false

	vblInterrupt bool // The VBL interrupt fired, until $C07x is accessed
	xInterrupt   bool // An X0 edge interrupt fired, until $C048 is accessed
	yInterrupt   bool // An Y0 edge interrupt fired, until $C048 is accessed
	xyIRQ        bool // The X0 or Y0 interrupt asserts the IRQ line, until read

	x1 bool // Mouse X direction, right when set
	y1 bool // Mouse Y direction, up when set

	mouseX mouseAxis
	mouseY mouseAxis

	frame uint64 // Last screen frame, the VBL interrupt fires on the next
}

/*
The mouse of the IIc has no counters, the firmware counts the steps. Each step
on an axis is an edge on X0 or Y0 that interrupts the CPU, and the firmware
reads X1 or Y1 to know the direction. A step is a full cycle of X0 or Y0, with
one edge of the kind the firmware chose to be interrupted on.

The frontends give where the pointer is on the window. Its moves become steps,
with the window as wide and high as the double hires screen, 560 by 192.
Programs clamp the mouse to the screen, and then the pointer of the IIc moves
as the one of the host. They drift apart when the program clamps to a smaller
range, places the mouse or misses steps. To meet again, the pointer of the host
pushes while it is on an edge of the window: there are steps towards the edge
every frame, and the program clamps them.

The steps are not lost when the firmware is busy: the next one on an axis waits
for the interrupt of the last to be served. They are also spaced to leave time
to the program, the firmware takes about 300 cycles to serve each one.
*/
const (
	mouseStepsX     = 560
	mouseStepsY     = 192
	mouseStepCycles = 1000 // Minimum between steps on an axis
	mouseEdge       = 1024 // Band of the host range that is the edge
	mouseEdgePush   = 16   // Steps towards the edge each frame on it
)

// mouseAxis turns the moves of the host pointer on an axis into steps
type mouseAxis struct {
	steps    int    // Steps the window is wide or high
	position int    // Position of the host pointer, in steps
	pending  int    // Steps to do, the sign is the direction
	lastStep uint64 // Cycle of the last step
}

// moveTo moves the host pointer, coming from the full range 0 to 65535. It
// is called once a frame.
func (m *mouseAxis) moveTo(position uint16) {
	target := int(position) * m.steps / 65536
	m.pending += target - m.position
	m.position = target

	// On the edge, push towards it once the moves are done
	if m.pending == 0 {
		if position < mouseEdge {
			m.pending = -mouseEdgePush
		} else if position > 65535-mouseEdge {
			m.pending = mouseEdgePush
		}
	}
}

// step returns whether there is a step to do now, and its direction. The
// step waits while the last one is busy.
func (m *mouseAxis) step(cycles uint64, busy bool) (stepped bool, positive bool) {
	if m.pending == 0 || busy || cycles < m.lastStep+mouseStepCycles {
		return false, false
	}
	m.lastStep = cycles
	positive = m.pending > 0
	if positive {
		m.pending--
	} else {
		m.pending++
	}
	return true, positive
}

func newIouApple2c(a *Apple2) *iouApple2c {
	return &iouApple2c{
		a:                   a,
		iouSwitchesDisabled: true,
		mouseX:              mouseAxis{steps: mouseStepsX},
		mouseY:              mouseAxis{steps: mouseStepsY},
	}
}

// tick fires the VBL interrupt when the vertical blank starts and moves the
// mouse
func (iou *iouApple2c) tick() {
	cycles := iou.a.GetCycles()
	frame := cycles / screenDrawCycles
	if frame != iou.frame {
		iou.frame = frame
		if iou.vblMask {
			iou.vblInterrupt = true
			iou.updateIRQ()
		}
		iou.readMouse()
	}

	if stepped, right := iou.mouseX.step(cycles, iou.xInterrupt); stepped {
		iou.x1 = right
		iou.mouseStepped(&iou.xInterrupt)
	}
	if stepped, down := iou.mouseY.step(cycles, iou.yInterrupt); stepped {
		iou.y1 = !down
		iou.mouseStepped(&iou.yInterrupt)
	}
}

// readMouse gets the moves of the host pointer, once a frame
func (iou *iouApple2c) readMouse() {
	if iou.a.io.mouse == nil {
		return
	}
	x, y, _ := iou.a.io.mouse.ReadMouse()
	iou.mouseX.moveTo(x)
	iou.mouseY.moveTo(y)
}

// mouseStepped interrupts for the step on an axis, if the movement
// interrupts are enabled. If not, the step is lost.
func (iou *iouApple2c) mouseStepped(interrupt *bool) {
	if iou.xyMask {
		*interrupt = true
		iou.xyIRQ = true
		iou.updateIRQ()
	}
}

func (iou *iouApple2c) updateIRQ() {
	iou.a.requestIRQ(irqSourceIOU, iou.vblInterrupt || iou.xyIRQ)
}

func (iou *iouApple2c) clearVBLInterrupt() {
	iou.vblInterrupt = false
	iou.updateIRQ()
}

func (iou *iouApple2c) clearXYInterrupts() {
	iou.xInterrupt = false
	iou.yInterrupt = false
	iou.xyIRQ = false
	iou.updateIRQ()
}

// lowerXYIRQ releases the IRQ line of the mouse movement, the flags stay
func (iou *iouApple2c) lowerXYIRQ() {
	iou.xyIRQ = false
	iou.updateIRQ()
}

func addApple2cSoftSwitches(io *ioC0Page) {
	a := io.apple2
	iou := newIouApple2c(a)
	a.registerTicker(iou)
	a.usesMouse = true

	// The ROM is always on $C100 to $CFFF, INTCXROM can't be changed
	a.mmu.intCxROMActive = true
	io.addSoftSwitchW(0x06, buildNotImplementedSoftSwitchW(io), "INTCXROMOFF-NONE")
	io.addSoftSwitchW(0x07, buildNotImplementedSoftSwitchW(io), "INTCXROMON-NONE")

	// The status of INTCXROM and SLOTC3ROM are replaced by the mouse ones.
	// Reading them does not reset the interrupts, the ROM reads them twice.
	io.addSoftSwitchR(0x15, func() uint8 {
		iou.lowerXYIRQ()
		return ssFromBool(iou.xInterrupt)
	}, "RSTXINT")
	io.addSoftSwitchR(0x17, func() uint8 {
		iou.lowerXYIRQ()
		return ssFromBool(iou.yInterrupt)
	}, "RSTYINT")

	// It is not the vertical blank but the VBL interrupt. See IIc Technical
	// Note #9.
	io.addSoftSwitchR(0x19, func() uint8 {
		return ssFromBool(iou.vblInterrupt)
	}, "RDVBLBAR")

	io.addSoftSwitchR(0x40, func() uint8 { return ssFromBool(iou.xyMask) }, "RDXYMSK")
	io.addSoftSwitchR(0x41, func() uint8 { return ssFromBool(iou.vblMask) }, "RDVBLMSK")
	io.addSoftSwitchR(0x42, func() uint8 { return ssFromBool(iou.x0Falling) }, "RDX0EDGE")
	io.addSoftSwitchR(0x43, func() uint8 { return ssFromBool(iou.y0Falling) }, "RDY0EDGE")
	io.addSoftSwitchRW(0x48, func() uint8 {
		iou.clearXYInterrupts()
		return 0
	}, "RSTXY")

	// With IOUDIS off, $C058 to $C05F are the mouse switches. With it on, they
	// are the annunciators as on the IIe, $C05E and $C05F the double hires.
	iouSwitches := []struct {
		name   string
		action func()
	}{
		{"DISXY", func() { iou.xyMask = false }},
		{"ENBXY", func() { iou.xyMask = true }},
		{"DISVBL", func() {
			iou.vblMask = false
			iou.clearVBLInterrupt()
		}},
		{"ENVBL", func() { iou.vblMask = true }},
		{"X0EDGE", func() { iou.x0Falling = false }},
		{"FALX0EDGE", func() { iou.x0Falling = true }},
		{"Y0EDGE", func() { iou.y0Falling = false }},
		{"FALY0EDGE", func() { iou.y0Falling = true }},
	}
	for i, sw := range iouSwitches {
		address := uint8(0x58 + i)
		annunciator := io.softSwitchesR[address]
		name := sw.name + "-" + io.softSwitchesRName[address]
		action := sw.action
		io.addSoftSwitchRW(address, func() uint8 {
			if iou.iouSwitchesDisabled {
				return annunciator()
			}
			action()
			return 0
		}, name)
	}

	// The 80/40 switch, up for 80 columns. It is down.
	io.addSoftSwitchR(0x60, func() uint8 { return ssOn }, "RD80SW")

	// The mouse button replaces PB2, pressed when bit 7 is zero
	mouseButton := func() uint8 {
		if io.mouse != nil {
			if _, _, pressed := io.mouse.ReadMouse(); pressed {
				return ssOff
			}
		}
		return ssOn
	}
	io.addSoftSwitchR(0x63, mouseButton, "RDMOUBTN")
	io.addSoftSwitchR(0x6b, mouseButton, "RDMOUBTN")

	// The mouse directions replace the paddles 2 and 3
	mouseX1 := func() uint8 { return ssFromBool(iou.x1) }
	mouseY1 := func() uint8 { return ssFromBool(iou.y1) }
	io.addSoftSwitchR(0x66, mouseX1, "MOUX1")
	io.addSoftSwitchR(0x6e, mouseX1, "MOUX1")
	io.addSoftSwitchR(0x67, mouseY1, "MOUY1")
	io.addSoftSwitchR(0x6f, mouseY1, "MOUY1")

	/*
		Any access to $C070 to $C07F triggers the paddles and resets the VBL
		interrupt. The ones from $C078 set IOUDIS on the even addresses and
		clear it on the odd ones, and read it on the even ones. The odd ones
		read the double hires, on when bit 7 is zero.
	*/
	strobePaddles := buildStrobePaddlesSoftSwitch(io)
	for address := uint8(0x70); address <= 0x7f; address++ {
		isIoudisSwitch := address >= 0x78
		isOdd := address&1 == 1
		io.addSoftSwitchR(address, func() uint8 {
			value := strobePaddles()
			iou.clearVBLInterrupt()
			if isIoudisSwitch {
				if isOdd {
					// Annunciator 3 is off with the double hires
					return io.softSwitchesData[ioFlagAnnunciator3]
				}
				return ssFromBool(iou.iouSwitchesDisabled)
			}
			return value
		}, "PTRIG")
		io.addSoftSwitchW(address, func(uint8) {
			strobePaddles()
			iou.clearVBLInterrupt()
			if isIoudisSwitch {
				iou.iouSwitchesDisabled = !isOdd
			}
		}, "PTRIG")
	}
	io.softSwitchesRName[0x78] = "RDIOUDIS"
	io.softSwitchesRName[0x79] = "RDDHIRES"
	io.softSwitchesWName[0x78] = "IOUDISON"
	io.softSwitchesWName[0x79] = "IOUDISOFF"
}

/*
apple2cBuiltInCards are the cards the IIc board is made of, on the slots their
firmware expects. They are the only ones allowed, the IIc has no slots.
*/
var apple2cBuiltInCards = map[int][]string{
	0: {"language"},
	6: {"diskii", "diskiiseq"},
}

// checkApple2cConfiguration refuses what can't be added to a IIc
func checkApple2cConfiguration(configuration *configuration) error {
	for slot := range 8 {
		cardConfig := configuration.get(fmt.Sprintf("s%v", slot))
		cardName := splitConfigurationString(cardConfig, ',')[0]
		if cardName == "" || cardName == noCardName {
			if slot == 0 {
				return fmt.Errorf("the language card is built in the Apple IIc, s0 must be 'language'")
			}
			continue
		}
		allowed := false
		for _, name := range apple2cBuiltInCards[slot] {
			allowed = allowed || cardName == name
		}
		if !allowed {
			return fmt.Errorf("the Apple IIc has no slots, '%s' can't be added on slot %v", cardName, slot)
		}
	}

	for _, key := range []string{confRamworks, confTape} {
		if value := configuration.get(key); value != "" && value != "none" {
			return fmt.Errorf("%s can't be used with the Apple IIc", key)
		}
	}
	for _, key := range []string{confRgb, confRomx} {
		if configuration.getFlag(key) {
			return fmt.Errorf("%s can't be used with the Apple IIc", key)
		}
	}
	return nil
}
