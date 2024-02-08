package izapple2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ivanizag/iz6502"
)

func configure(configuration *configuration) (*Apple2, error) {
	a := newApple2()
	a.Name = configuration.get(confName)

	// Configure the board
	board := configuration.get(confBoard)
	a.board = board
	a.isApple2e = board == "2e"

	addApple2SoftSwitches(a.io)
	if a.isApple2e {
		a.mmu.initExtendedRAM(1)
		addApple2ESoftSwitches(a.io)
	}
	if board == "base64a" {
		addBase64aSoftSwitches(a.io)
	}

	cpu := configuration.get(confCpu)
	switch cpu {
	case "6502":
		a.cpu = iz6502.NewNMOS6502(a.mmu)
	case "65c02":
		a.cpu = iz6502.NewCMOS65c02(a.mmu)
	}

	err := a.loadRom(configuration.get(confRom))
	if err != nil {
		return nil, err
	}

	err = setupCharactedGenerator(a, board, configuration.get(confCharRom))
	if err != nil {
		return nil, err
	}

	a.setProfiling(configuration.getFlag(confProfile))
	a.SetForceCaps(configuration.getFlag(confForceCaps))

	err = a.setClockSpeed(configuration.get(confSpeed))
	if err != nil {
		return nil, err
	}

	// Add cards on the slots
	for i := 0; i < 8; i++ {
		cardConfig := configuration.get(fmt.Sprintf("s%v", i))
		if cardConfig != "" {
			_, err := setupCard(a, i, cardConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	// Add mods
	mods := strings.Split(configuration.get(confMods), ",")
	for _, mod := range mods {
		switch strings.TrimSpace(mod) {
		//case "shift":
		//	setupShiftedKeyboard(a)
		case "four-colors":
			// This removes the mod to have 6 colors sent by Wozniak to Byte
			// magazine. See: https://archive.org/details/byte-magazine-1979-06/page/n67/mode/2up?view=theater
			a.isFourColors = true
		}
	}

	// Add optional accesories including the aux slot
	ramWorksSize := configuration.get(confRamworks)
	if ramWorksSize != "" && ramWorksSize != "none" {
		err = setupRAMWorksCard(a, ramWorksSize)
		if err != nil {
			return nil, err
		}
	}

	if configuration.getFlag(confRgb) {
		setupRGBCard(a)
	}

	nsc := configuration.get(confNsc)
	if nsc != "none" && nsc != "" {
		err = setupNoSlotClock(a, nsc)
		if err != nil {
			return nil, err
		}
	}

	if configuration.getFlag(confRomx) {
		err := setupRomX(a)
		if err != nil {
			return nil, err
		}
	}

	err = setupTracers(a, configuration.get(confTrace))
	if err != nil {
		return nil, err
	}

	return a, nil
}

func newApple2() *Apple2 {
	var a Apple2

	a.Name = "Pending"
	a.mmu = newMemoryManager(&a)
	a.io = newIoC0Page(&a)
	a.commandChannel = make(chan command, 100)

	return &a
}

func (a *Apple2) setClockSpeed(speed string) error {
	if speed == "full" {
		a.cycleDurationNs = 0
	} else if speed == "ntsc" {
		a.cycleDurationNs = 1000.0 / CPUClockMhz
	} else if speed == "pal" {
		a.cycleDurationNs = 1000.0 / cpuClockEuroMhz
	} else {
		clockMhz, err := strconv.ParseFloat(speed, 64)
		if err != nil {
			return fmt.Errorf("invalid clock speed: %s", speed)
		}
		a.cycleDurationNs = 1000.0 / clockMhz
	}
	return nil
}

func (a *Apple2) setProfiling(value bool) {
	a.profile = value
}

// SetForceCaps allows the caps state to be toggled at runtime
func (a *Apple2) SetForceCaps(value bool) {
	a.forceCaps = value
}

func (a *Apple2) loadRom(filename string) error {
	if a.board == "base64a" && filename == "<custom>" {
		// The ROM of the base64a has several file and pages
		loadBase64aRom(a)
		return nil
	}

	data, _, err := LoadResource(filename)
	if err != nil {
		return err
	}

	size := len(data)

	romBase := 0x10000 - size
	a.mmu.physicalROM = newMemoryRangeROM(uint16(romBase), data, "Main ROM")
	return nil
}

// CreateConfiguredApple is a device independent main. Video, keyboard and speaker won't be defined
func CreateConfiguredApple() (*Apple2, error) {
	// Get configuration from defaults and the command line
	configuration, filename, err := getConfigurationFromCommandLine()
	if err != nil {
		return nil, err
	}

	if filename != "" {
		// Try loading as diskette
		_, err := LoadDiskette(filename)
		isDiskette := err == nil
		if isDiskette {
			// Let's force a DiskII with the diskette in slot 6
			configuration.set(confS6, fmt.Sprintf("diskii,disk1=\"%s\"", filename))
		} else {
			// Let's force a Smartport card with a block device in slot 7
			configuration.set(confS7, fmt.Sprintf("smartport,image1=\"%s\"", filename))
		}
	}

	a, err := configure(configuration)
	if err != nil {
		return nil, err
	}
	return a, nil
}
