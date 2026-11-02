package izapple2

import (
	"fmt"
	"strconv"

	"github.com/ivanizag/izapple2/component"
)

/*
The disk port of the Apple IIc, an IWM on the I/O addresses of the slot 6,
with the internal 5.25 drive as drive 1 and the external port as drive 2.
Its firmware is on the ROM of the IIc: the 5.25 drives on $C600, and from
the ROM 0 the SmartPort on $C500.

See:

	"Apple IIc Technical Reference Manual", Chapter 11, The disk
	controller unit, and Appendix I, the firmware listing.
	"Integrated Woz Machine (IWM) Specification", Apple, rev 19, 1982.

The 5.25 drives are emulated by a Disk II card, as diskii or diskiiseq,
and the devices of the external port by a SmartPort bus, as on the Liron
card. Each access goes to the IWM, that has the mode, the status and the
handshake registers:

  - In the synchronous mode, the mode of the 5.25 drives, the access goes
    too to the Disk II. Its result is the one of the Disk II, but for the
    status register, with the SENSE of the Disk II, and the handshake
    register.
  - In the asynchronous mode, the mode of the SmartPort, only the phases,
    the motor and the drive selection go to the Disk II. Its data
    accesses would write the packets on a diskette of the drive 2.

The bus sees the phases of the 5.25 drives, it only takes the phases 1
and 3 together, or 0 and 2, that the drives never use to step.
*/
type CardDiskIIc struct {
	cardBase
	disk2      Card
	disk2Read  [16]softSwitchR // The soft switches of the Disk II
	disk2Write [16]softSwitchW
	iwm        *component.IWM
	bus        *smartPortBus
}

const (
	iicDiskMaxUnits = 4

	iwmModeAsynchronous uint8 = 1 << 1
	iwmStateL6          uint8 = 1 << 6
	iwmStateL7          uint8 = 1 << 7
)

func newCardDiskIIcBuilder() *cardBuilder {
	params := []paramSpec{
		{"disk1", "Diskette image for the internal drive", ""},
		{"disk2", "Diskette image for the external 5.25 drive", ""},
		{"sequencer", "Emulate the 5.25 drives as diskiiseq, with the Disk II state machine", "false"},
		{"fast", "Enable CPU burst when accessing the disk", "true"},
		{"tracktracer", "Trace how the disk head moves between tracks", "false"},
	}
	for i := 1; i <= iicDiskMaxUnits; i++ {
		params = append(params, paramSpec{"image" + strconv.Itoa(i), "Disk image for the SmartPort unit " + strconv.Itoa(i), ""})
	}
	params = append(params,
		paramSpec{"tracesp", "Trace the SmartPort bus", "false"},
		paramSpec{"tracehd", "Trace image accesses", "false"},
		saveDirParamSpec,
	)
	return &cardBuilder{
		name:          "Apple IIc disk port",
		description:   "Disk port of the Apple IIc, an IWM with the 5.25 drives and a SmartPort",
		defaultParams: &params,
		requiresIIc:   true,
		hide:          true,
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardDiskIIc
			err := c.buildDisk2(params)
			if err != nil {
				return nil, err
			}

			c.bus = newSmartPortBus()
			c.bus.trace = paramsGetBool(params, "tracesp")
			traceHD := paramsGetBool(params, "tracehd")
			for i := 1; i <= iicDiskMaxUnits; i++ {
				image := paramsGetPath(params, "image"+strconv.Itoa(i))
				if image != "" {
					device, err := NewSmartPortHardDisk(image, paramsGetSaveDir(params))
					if err != nil {
						return nil, err
					}
					device.trace = traceHD
					c.bus.addDevice(device)
				}
			}
			c.iwm = component.NewIWM(c.bus)
			return &c, nil
		},
	}
}

// buildDisk2 builds the Disk II card for the 5.25 drives with the parameters
// that apply to it
func (c *CardDiskIIc) buildDisk2(params map[string]string) error {
	name := "diskii"
	if paramsGetBool(params, "sequencer") {
		name = "diskiiseq"
	}
	builder := getCardFactory()[name]
	disk2Params := builder.fullDefaultParams()
	for key := range disk2Params {
		if value, ok := params[key]; ok {
			disk2Params[key] = value
		}
	}
	card, err := builder.buildFunc(disk2Params)
	if err != nil {
		return err
	}
	card.configure(builder.name, false, false)

	if disk2CardBase(card) == nil {
		return fmt.Errorf("unexpected card %T for the 5.25 drives", card)
	}
	c.disk2 = card
	return nil
}

// disk2CardBase is the cardBase of the Disk II cards, where they keep their
// soft switches
func disk2CardBase(card Card) *cardBase {
	switch disk2 := card.(type) {
	case *CardDisk2:
		return &disk2.cardBase
	case *CardDisk2Sequencer:
		return &disk2.cardBase
	default:
		return nil
	}
}

func (c *CardDiskIIc) assign(a *Apple2, slot int) {
	// The Disk II registers its soft switches, they are kept to call them and
	// replaced below
	c.disk2.assign(a, slot)
	base := disk2CardBase(c.disk2)
	c.disk2Read = base._ssr
	c.disk2Write = base._ssw

	for address := range uint8(16) {
		c.addCardSoftSwitchR(address, func() uint8 {
			return c.read(address)
		}, "IWM")
		c.addCardSoftSwitchW(address, func(value uint8) {
			c.write(address, value)
		}, "IWM")
	}
	c.cardBase.assign(a, slot)
}

func (c *CardDiskIIc) isSynchronous() bool {
	return c.iwm.Mode()&iwmModeAsynchronous == 0
}

func (c *CardDiskIIc) read(address uint8) uint8 {
	value := c.iwm.Access(c.a.GetCycles(), address, false, 0)
	if !c.isSynchronous() {
		if address < 0xc {
			c.disk2Read[address]()
		}
		return value
	}

	disk2Value := c.disk2Read[address]()
	if address&1 == 0 {
		state := c.iwm.State()
		l6 := state&iwmStateL6 != 0
		l7 := state&iwmStateL7 != 0
		if l6 && !l7 {
			// The status, with the SENSE of the 5.25 drive
			return value&0x7f | disk2Value&0x80
		}
		if l7 && !l6 {
			// The handshake
			return value
		}
	}
	return disk2Value
}

func (c *CardDiskIIc) write(address uint8, value uint8) {
	c.iwm.Access(c.a.GetCycles(), address, true, value)
	if c.isSynchronous() || address < 0xc {
		c.disk2Write[address](value)
	}
}

func (c *CardDiskIIc) reset() {
	c.iwm.Reset()
	c.disk2.reset()
}

// GetInfo returns the info of the 5.25 drives
func (c *CardDiskIIc) GetInfo() map[string]string {
	return c.disk2.GetInfo()
}
