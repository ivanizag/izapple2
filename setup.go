package izapple2

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ivanizag/iz6502"
)

func configure(configuration *configuration) (*Apple2, error) {

	if configuration.getFlag(confShowConfig) {
		configuration.dump()
		os.Exit(0)
	}

	var a Apple2
	a.Name = configuration.get(confName)
	a.mmu = newMemoryManager(&a)
	a.video = newVideo(&a)
	a.io = newIoC0Page(&a)
	a.commandChannel = make(chan command, 100)

	// Configure the board
	board := configuration.get(confBoard)
	a.board = board

	err := setupCharactedGenerator(&a, board, configuration.get(confCharRom))
	if err != nil {
		return nil, err
	}

	addApple2SoftSwitches(a.io)
	switch board {
	case "2plus":
		a.mmu.initMainRAM()
	case "2e":
		a.isApple2e = true
		a.mmu.initMainRAM()
		a.mmu.initExtendedRAM(1)
		a.hasLowerCase = true
		addApple2ESoftSwitches(a.io)
	case "base64a":
		a.mmu.initMainRAM()
		a.hasLowerCase = true
		addBase64aSoftSwitches(a.io)
	case "basis108":
		memBasis108 := newMemoryRangeBasis108()
		videoBasis108 := newVideoBasis108(&a, memBasis108)
		a.mmu.initCustomRAM(memBasis108)
		a.video = videoBasis108
		a.hasLowerCase = true
		addBasis108SoftSwitches(a.io, memBasis108, videoBasis108, a.cg)
	default:
		return nil, fmt.Errorf("board %s not supported it must be '2plus', '2e', 'base64a', 'basis108", board)
	}

	cpu := configuration.get(confCpu)
	switch cpu {
	case "6502":
		a.cpu = iz6502.NewNMOS6502(a.mmu)
	case "65c02":
		a.cpu = iz6502.NewCMOS65c02(a.mmu)
	}

	err = a.loadRom(configuration.get(confRom))
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
			_, err := setupCard(&a, i, cardConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	// Add mods
	mods := strings.Split(configuration.get(confMods), ",")
	for _, mod := range mods {
		switch strings.TrimSpace(mod) {
		// case "shift":
		//	 setupShiftedKeyboard(a)
		case "four-colors":
			// This removes the mod to have 6 colors sent by Wozniak to Byte
			// magazine. See: https://archive.org/details/byte-magazine-1979-06/page/n67/mode/2up?view=theater
			a.isFourColors = true
		}
	}

	// Add optional accesories including the aux slot
	ramWorksSize := configuration.get(confRamworks)
	if ramWorksSize != "" && ramWorksSize != "none" {
		err = setupRAMWorksCard(&a, ramWorksSize)
		if err != nil {
			return nil, err
		}
	}

	if configuration.getFlag(confRgb) {
		setupRGBCard(&a)
	}

	nsc := configuration.get(confNsc)
	if nsc != "none" && nsc != "" {
		err = setupNoSlotClock(&a, nsc)
		if err != nil {
			return nil, err
		}
	}

	if configuration.getFlag(confRomx) {
		err := setupRomX(&a)
		if err != nil {
			return nil, err
		}
	}

	err = setupTracers(&a, configuration.get(confTrace))
	if err != nil {
		return nil, err
	}

	return &a, nil
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
	if filename == "<custom>" {
		switch a.board {
		case "base64a":
			return loadBase64aRom(a)
		case "basis108":
			return loadBasis108Rom(a)
		default:
			return fmt.Errorf("no custom ROM defined for board %s", a.board)
		}
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

const pagedRomChipWindowSize = 0x800                                  // 2 KB
const pagedRomChipCount = 6                                           // There has to be six ROM chips
const pagedRomWindowSize = pagedRomChipWindowSize * pagedRomChipCount // To cover 0xd000 to 0xffff
func loadMultiPageRom(a *Apple2, filenames []string) error {
	if len(filenames) != pagedRomChipCount {
		return fmt.Errorf("expected %d ROM files, got %d", pagedRomChipCount, len(filenames))
	}

	// Load the 6 PROM dumps
	proms := make([][]uint8, pagedRomChipCount)
	banks := 1
	for i, filename := range filenames {
		var err error
		proms[i], _, err = LoadResource(filename)
		if err != nil {
			return err
		}
		pages := len(proms[i]) / pagedRomChipWindowSize
		if pages > banks {
			banks = pages
		}
	}

	// Init the array of banks
	romBanksBytes := make([][]uint8, banks)
	for bank := range romBanksBytes {
		romBanksBytes[bank] = make([]uint8, 0, pagedRomWindowSize)
	}

	// Distribute the per chip banks on the full rom banks
	for _, romData := range proms {
		for bank := range romBanksBytes {
			start := (bank * pagedRomChipWindowSize) % len(romData)
			romBanksBytes[bank] = append(romBanksBytes[bank], romData[start:start+pagedRomChipWindowSize]...)
		}
	}

	// Create paged ROM
	romData := make([]uint8, 0, pagedRomWindowSize*banks)
	for _, bank := range romBanksBytes {
		romData = append(romData, bank...)
	}
	rom := newMemoryRangePagedROM(0xd000, romData, "Multipage main ROM", uint8(banks))

	// Start with first bank active
	rom.setPage(0)

	a.mmu.physicalROM = rom
	return nil
}

// CreateConfiguredApple is a device independent main. Video, keyboard and speaker won't be defined
func CreateConfiguredApple() (*Apple2, error) {
	// Get configuration from defaults and the command line
	configuration, filenames, err := getConfigurationFromCommandLine()
	if err != nil {
		return nil, err
	}

	if len(filenames) > 0 {
		diskettes := []string{}
		blockDevices := []string{}
		for _, filename := range filenames {
			_, err := LoadDiskette(filename)
			isDiskette := err == nil
			if isDiskette {
				diskettes = append(diskettes, filename)
			} else {
				blockDevices = append(blockDevices, filename)
			}
		}

		if len(diskettes) == 1 {
			configuration.set(confS6, fmt.Sprintf("diskii,disk1=\"%s\"", filenames[0]))
		} else if len(diskettes) >= 2 {
			configuration.set(confS6, fmt.Sprintf("diskii,disk1=\"%s\",disk2=\"%s\"", filenames[0], filenames[1]))
		}
		if len(diskettes) == 3 {
			configuration.set(confS5, fmt.Sprintf("diskii,disk1=\"%s\"", filenames[2]))
		} else if len(diskettes) >= 4 {
			configuration.set(confS5, fmt.Sprintf("diskii,disk1=\"%s\",disk2=\"%s\"", filenames[2], filenames[3]))
		}
		if len(diskettes) > 4 {
			return nil, fmt.Errorf("up to 4 diskettes can be loaded, %v found", len(diskettes))
		}

		if len(blockDevices) > 8 {
			return nil, fmt.Errorf("up to 8 block devices can be loaded, %v found", len(blockDevices))
		}
		if len(blockDevices) > 0 {
			configuration.set(confS7, fmt.Sprintf("smartport,image1=\"%s\"", blockDevices[0]))
			if len(blockDevices) > 1 {
				smartportConfig := "smartport"
				for i, filename := range blockDevices {
					if i == 0 {
						continue
					}
					smartportConfig += fmt.Sprintf(",image%v=\"%s\"", i+1, filename)
				}
				configuration.set(confS5, smartportConfig)
			}
		}
	}

	a, err := configure(configuration)
	if err != nil {
		return nil, err
	}
	return a, nil
}
