package izapple2

import (
	"strconv"

	"github.com/ivanizag/izapple2/component"
)

/*
Apple II UniDisk 3.5 interface card, known as Liron. An IWM driving a
SmartPort bus, with the firmware of Apple on its ROM.

See:

	The firmware disassembled by Eric Smith:
	https://git.applefritter.com/Apple-2-SW/a2-liron
	The ROM dump of the Yellowstone project by Steve Chamberlin:
	https://git.applefritter.com/zjkuki/fpga-disk-controller

The ROM has 4 KB for the whole $Cxxx area: a page for each slot, the same
but for the slot number, and the shared code on $C800.

The IWM is on the sixteen I/O addresses of the slot. The devices on the bus
are SmartPort hard disks with the images given, as the UniDisk 3.5 drives
and the SmartPort hard disks. The firmware of the drives is not emulated.

With the fast option, the emulation runs at full speed while the firmware
uses the card. The bus and the motor stay enabled after a call, but the
firmware only accesses the IWM while it runs one: the fast mode is released
some time after the last access.
*/
type CardLiron struct {
	cardBase
	iwm *component.IWM
	bus *smartPortBus

	fastMode   bool
	fast       bool   // Fast mode requested
	lastAccess uint64 // Cycle of the last access to the IWM
}

const (
	lironMaxUnits = 4

	// Cycles without accesses to the IWM to release the fast mode, a tenth
	// of a second, longer than the waits of the firmware in a call
	lironFastHoldCycles = 102_000
)

func newCardLironBuilder() *cardBuilder {
	params := []paramSpec{}
	for i := 1; i <= lironMaxUnits; i++ {
		params = append(params, paramSpec{"image" + strconv.Itoa(i), "Disk image for unit " + strconv.Itoa(i), ""})
	}
	params = append(params,
		paramSpec{"fast", "Enable CPU burst when accessing the disk", "true"},
		paramSpec{"tracesp", "Trace the SmartPort bus", "false"},
		paramSpec{"tracehd", "Trace image accesses", "false"},
		saveDirParamSpec,
	)
	return &cardBuilder{
		name:          "UniDisk 3.5 interface",
		description:   "Apple UniDisk 3.5 interface card, Liron, a SmartPort on an IWM",
		defaultParams: &params,
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardLiron
			c.bus = newSmartPortBus()
			c.bus.trace = paramsGetBool(params, "tracesp")
			traceHD := paramsGetBool(params, "tracehd")
			saveDirectory := paramsGetSaveDir(params)
			for i := 1; i <= lironMaxUnits; i++ {
				image := paramsGetPath(params, "image"+strconv.Itoa(i))
				if image != "" {
					device, err := NewSmartPortHardDisk(image, saveDirectory)
					if err != nil {
						return nil, err
					}
					device.trace = traceHD
					c.bus.addDevice(device)
				}
			}
			c.iwm = component.NewIWM(c.bus)
			c.fastMode = paramsGetBool(params, "fast")

			err := c.loadRomFromResource("<internal>/Apple II UniDisk 3.5 Interface Card ROM.bin", cardRomFull)
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
	}
}

func (c *CardLiron) assign(a *Apple2, slot int) {
	for address := range uint8(16) {
		c.addCardSoftSwitchR(address, func() uint8 {
			c.accessed()
			return c.iwm.Access(a.GetCycles(), address, false, 0)
		}, "IWM")
		c.addCardSoftSwitchW(address, func(value uint8) {
			c.accessed()
			c.iwm.Access(a.GetCycles(), address, true, value)
		}, "IWM")
	}
	if c.fastMode {
		a.registerTicker(c)
	}
	c.cardBase.assign(a, slot)
}

// accessed starts or extends the fast mode on an access to the IWM
func (c *CardLiron) accessed() {
	if !c.fastMode {
		return
	}
	c.lastAccess = c.a.GetCycles()
	if !c.fast {
		c.fast = true
		c.a.RequestFastMode()
		c.a.requestTickAt(c.lastAccess + lironFastHoldCycles)
	}
}

// tick releases the fast mode when the IWM has not been accessed for a while
func (c *CardLiron) tick() uint64 {
	if !c.fast {
		return tickNever
	}
	release := c.lastAccess + lironFastHoldCycles
	if c.a.GetCycles() < release {
		return release
	}
	c.fast = false
	c.a.ReleaseFastMode()
	return tickNever
}

func (c *CardLiron) reset() {
	c.iwm.Reset()
}
