package izapple2

import (
	"bytes"
	"os"
	"testing"

	"github.com/ivanizag/izapple2/fujinet"
)

const (
	spTestParams  = uint16(0x0340) // Parameter list of the calls
	spTestControl = uint16(0x0360) // Control list
	spTestResult  = uint16(0x0350) // Result in A, then X and Y
	spTestBuffer  = uint16(0x4000)
	spTestBuffer2 = uint16(0x4400)
)

// smartPortMachine starts a //e with DOS 3.3 and a card on slot 5
func smartPortMachine(t *testing.T, card string) *Apple2 {
	t.Helper()
	a, err := CreateAppleFromModel("2enh", map[string]string{
		"s5":        card,
		"startTime": "1985-03-15T10:30:00",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	a.Init()
	cyclesToText(t, a, "DOS VERSION 3.3", 100_000_000)
	a.RunCycles(1_000_000)
	return a
}

// runSmartPortProgram runs a program at $0300 that ends with BRA *
func runSmartPortProgram(a *Apple2, program []uint8) {
	a.mmu.pokeRange(0x0300, program)
	a.cpu.SetPC(0x0300)
	a.RunCycles(500_000)
}

/*
callSmartPort makes a call to the SmartPort entry of the card on a slot, as
the programs do: JSR to $Cn00 + ($CnFF) + 3, followed by the command and the
address of the parameter list. It returns the error code in A.
*/
func callSmartPort(a *Apple2, slot int, command uint8, params []uint8) uint8 {
	a.mmu.pokeRange(spTestParams, params)
	base := uint16(0xc000 + slot*0x100)
	dispatch := base + uint16(a.mmu.Peek(base+0xff)) + 3
	runSmartPortProgram(a, []uint8{
		0x20, uint8(dispatch), uint8(dispatch >> 8), // JSR dispatch
		command,
		lo(spTestParams), hi(spTestParams),
		0x8d, lo(spTestResult), hi(spTestResult), // STA result
		0x80, 0xfe, // BRA *
	})
	return a.mmu.Peek(spTestResult)
}

// proDOSBlockCall makes a call to the ProDOS block device entry of a card
func proDOSBlockCall(a *Apple2, slot int, command uint8, buffer uint16, block uint16) uint8 {
	a.mmu.pokeRange(0x42, []uint8{
		command,
		uint8(slot << 4),
		uint8(buffer), uint8(buffer >> 8),
		uint8(block), uint8(block >> 8),
	})
	base := uint16(0xc000 + slot*0x100)
	entry := base + uint16(a.mmu.Peek(base+0xff))
	runSmartPortProgram(a, []uint8{
		0x20, uint8(entry), uint8(entry >> 8), // JSR entry
		0x8d, lo(spTestResult), hi(spTestResult), // STA result
		0x8e, lo(spTestResult + 1), hi(spTestResult + 1), // STX result+1
		0x8c, lo(spTestResult + 2), hi(spTestResult + 2), // STY result+2
		0x80, 0xfe, // BRA *
	})
	return a.mmu.Peek(spTestResult)
}

func lo(address uint16) uint8 { return uint8(address) }
func hi(address uint16) uint8 { return uint8(address >> 8) }

func peekRange(a *Apple2, address uint16, length int) []uint8 {
	data := make([]uint8, length)
	for i := range data {
		data[i] = a.mmu.Peek(address + uint16(i))
	}
	return data
}

func TestSmartPortBootsProDOS(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	overrides := newConfiguration()
	overrides.set(confS7, "smartport,image1=\""+image+"\"")
	testBoots(t, "2enh", "", overrides, 100_000_000, "BITSY  BYE", "NEW VOL", testTextMode40)
}

func TestSmartPortStatus(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	a := smartPortMachine(t, "smartport,image1=\""+image+"\"")

	// The host, with one device
	result := callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 0, lo(spTestBuffer), hi(spTestBuffer), 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error on the status of the host, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer, 8); !bytes.Equal(got, []uint8{1, 0xff, 0, 0, 1, 0, 0, 0}) {
		t.Errorf("unexpected status of the host % x", got)
	}

	// The hard disk
	result = callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 1, lo(spTestBuffer2), hi(spTestBuffer2), 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error on the status of the unit, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer2, 8); !bytes.Equal(got, []uint8{1, 0xff, 0, 0, 1, 0, 0, 0}) {
		t.Errorf("unexpected status of the unit % x", got)
	}

	// A unit that is not there. The status with code 0 of any unit is the
	// one of the host, as it was on the card
	result = callSmartPort(a, 5, smartPortCommandReadBlock, []uint8{3, 4, lo(spTestBuffer2), hi(spTestBuffer2), 0, 0, 0})
	if result != smartPortErrorNoDevice {
		t.Errorf("expected no device on the unit 4, got $%02x", result)
	}
}

func TestSmartPortReadAndWriteBlocks(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	original, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	a := smartPortMachine(t, "smartport,image1=\""+image+"\"")

	// Read the block 2
	result := callSmartPort(a, 5, smartPortCommandReadBlock, []uint8{3, 1, lo(spTestBuffer), hi(spTestBuffer), 2, 0, 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error reading, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer, 512); !bytes.Equal(got, original[2*512:3*512]) {
		t.Error("the block read is not the one on the image")
	}

	// Write the block 100 and read it back
	pattern := make([]uint8, 512)
	for i := range pattern {
		pattern[i] = uint8(i * 7)
	}
	a.mmu.pokeRange(spTestBuffer, pattern)
	result = callSmartPort(a, 5, smartPortCommandWriteBlock, []uint8{3, 1, lo(spTestBuffer), hi(spTestBuffer), 100, 0, 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error writing, got $%02x", result)
	}
	result = callSmartPort(a, 5, smartPortCommandReadBlock, []uint8{3, 1, lo(spTestBuffer2), hi(spTestBuffer2), 100, 0, 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error reading back, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer2, 512); !bytes.Equal(got, pattern) {
		t.Error("the block read back is not the one written")
	}
	written, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written[100*512:101*512], pattern) {
		t.Error("the block written is not on the image")
	}
}

func TestProDOSBlockReadsThroughTheProDOSEntry(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	original, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	a := smartPortMachine(t, "prodosblock,image1=\""+image+"\"")

	result := proDOSBlockCall(a, 5, smartPortCommandReadBlock, spTestBuffer, 3)
	if result != smartPortNoError {
		t.Fatalf("expected no error reading, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer, 512); !bytes.Equal(got, original[3*512:4*512]) {
		t.Error("the block read is not the one on the image")
	}

	// The status gives the size in blocks in X and Y
	result = proDOSBlockCall(a, 5, smartPortCommandStatus, spTestBuffer, 0)
	blocks := int(a.mmu.Peek(spTestResult+1)) + int(a.mmu.Peek(spTestResult+2))<<8
	if result != smartPortNoError || blocks != len(original)/512 {
		t.Errorf("expected the status with %v blocks, got $%02x and %v blocks", len(original)/512, result, blocks)
	}
}

func TestFujinetClockStatus(t *testing.T) {
	a := smartPortMachine(t, "fujinet")

	// The host, with the network and the clock
	callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 0, lo(spTestBuffer), hi(spTestBuffer), 0})
	if devices := a.mmu.Peek(spTestBuffer); devices != 2 {
		t.Errorf("expected two devices, got %v", devices)
	}

	// The time, 1985-03-15 10:30 and some seconds
	result := callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 2, lo(spTestBuffer), hi(spTestBuffer), 'T'})
	if result != smartPortNoError {
		t.Fatalf("expected no error, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer, 6); !bytes.Equal(got, []uint8{19, 85, 3, 15, 10, 30}) {
		t.Errorf("unexpected time % x", got)
	}

	// The same in the ProDOS format
	callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 2, lo(spTestBuffer), hi(spTestBuffer), 'P'})
	if got := peekRange(a, spTestBuffer, 4); !bytes.Equal(got, []uint8{15 + 3<<5, 85 << 1, 30, 10}) {
		t.Errorf("unexpected ProDOS time % x", got)
	}

	// The device information block
	callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 2, lo(spTestBuffer), hi(spTestBuffer), smartPortStatusCodeDeviceInfo})
	if got := peekRange(a, spTestBuffer+4, 9); !bytes.Equal(got, []uint8("\x08FN_CLOCK")) {
		t.Errorf("unexpected device information % x", got)
	}
}

func TestFujinetNetworkControlAndStatus(t *testing.T) {
	a := smartPortMachine(t, "fujinet")

	// The channel mode to JSON, with a control list of one byte
	a.mmu.pokeRange(spTestControl, []uint8{1, 0, 1})
	result := callSmartPort(a, 5, smartPortCommandControl, []uint8{3, 1, lo(spTestControl), hi(spTestControl), 0xfc})
	if result != smartPortNoError {
		t.Fatalf("expected no error on the control, got $%02x", result)
	}

	// In JSON mode with no data the status says end of file
	callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 1, lo(spTestBuffer), hi(spTestBuffer), 'S'})
	expected := []uint8{0, 0, 1, uint8(fujinet.NetworkErrorEndOfFile)}
	if got := peekRange(a, spTestBuffer, 4); !bytes.Equal(got, expected) {
		t.Errorf("expected the status % x in JSON mode, got % x", expected, got)
	}
}
