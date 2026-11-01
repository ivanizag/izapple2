package izapple2

import (
	"bytes"
	"os"
	"testing"
)

// The Liron calls go through the firmware of the card, the IWM and the bus

func TestLironBootsProDOS(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	overrides := newConfiguration()
	overrides.set(confS5, "liron,image1=\""+image+"\"")
	testBoots(t, "2enh", "", overrides, 100_000_000, "BITSY  BYE", "S5,D1:/PRODOS.2.4.3", testTextMode40)
}

func TestLironReadAndWriteBlocks(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	original, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	a := smartPortMachine(t, "liron,image1=\""+image+"\"")

	result := callSmartPort(a, 5, smartPortCommandReadBlock, []uint8{3, 1, lo(spTestBuffer), hi(spTestBuffer), 2, 0, 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error reading, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer, 512); !bytes.Equal(got, original[2*512:3*512]) {
		t.Error("the block read is not the one on the image")
	}

	// The block goes on a data packet after the command
	pattern := make([]uint8, 512)
	for i := range pattern {
		pattern[i] = uint8(i * 13)
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

func TestLironStatusOfTheDevice(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	a := smartPortMachine(t, "liron,image1=\""+image+"\"")

	result := callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 1, lo(spTestBuffer), hi(spTestBuffer), 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error, got $%02x", result)
	}
	// A block device for reading, writing and formatting, online, of 280 blocks
	if got := peekRange(a, spTestBuffer, 4); !bytes.Equal(got, []uint8{0xf8, 0x18, 0x01, 0x00}) {
		t.Errorf("unexpected status % x", got)
	}
}

func TestLironTwoUnits(t *testing.T) {
	first := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	second := writableCopy(t, "<internal>/cpm65.po")
	expected, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	a := smartPortMachine(t, "liron,image1=\""+first+"\",image2=\""+second+"\"")

	// The firmware numbers the units with INIT, the host status has two
	result := callSmartPort(a, 5, smartPortCommandStatus, []uint8{3, 0, lo(spTestBuffer), hi(spTestBuffer), 0})
	if result != smartPortNoError || a.mmu.Peek(spTestBuffer) != 2 {
		t.Errorf("expected two units, got $%02x and %v", result, a.mmu.Peek(spTestBuffer))
	}

	result = callSmartPort(a, 5, smartPortCommandReadBlock, []uint8{3, 2, lo(spTestBuffer), hi(spTestBuffer), 0, 0, 0})
	if result != smartPortNoError {
		t.Fatalf("expected no error reading the unit 2, got $%02x", result)
	}
	if got := peekRange(a, spTestBuffer, 512); !bytes.Equal(got, expected[:512]) {
		t.Error("the block read is not the one on the second image")
	}
}

func TestLironProDOSEntry(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	original, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	a := smartPortMachine(t, "liron,image1=\""+image+"\"")

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

func TestLironFastMode(t *testing.T) {
	image := writableCopy(t, "<internal>/ProDOS_2_4_3.po")
	a := smartPortMachine(t, "liron,image1=\""+image+"\"")
	// Other parts of the machine may be asking for the fast mode too
	before := a.fastRequestsCounter

	// Right after the call, the fast mode goes on a while
	readBlockUntilReturn(a, 5)
	a.RunCycles(lironFastHoldCycles / 2)
	if a.fastRequestsCounter != before+1 {
		t.Error("expected the fast mode after the call")
	}
	a.RunCycles(lironFastHoldCycles)
	if a.fastRequestsCounter != before {
		t.Error("expected the fast mode released after the call")
	}

	// Disabled
	a = smartPortMachine(t, "liron,fast=false,image1=\""+image+"\"")
	before = a.fastRequestsCounter
	readBlockUntilReturn(a, 5)
	if a.fastRequestsCounter != before {
		t.Error("expected no fast mode with fast=false")
	}
}

// readBlockUntilReturn runs a READBLOCK call to the instruction after it
func readBlockUntilReturn(a *Apple2, slot int) {
	a.mmu.pokeRange(spTestParams, []uint8{3, 1, lo(spTestBuffer), hi(spTestBuffer), 2, 0, 0})
	base := uint16(0xc000 + slot*0x100)
	dispatch := base + uint16(a.mmu.Peek(base+0xff)) + 3
	a.mmu.pokeRange(0x0300, []uint8{
		0x20, uint8(dispatch), uint8(dispatch >> 8), // JSR dispatch
		smartPortCommandReadBlock, lo(spTestParams), hi(spTestParams),
		0x80, 0xfe, // BRA *
	})
	a.cpu.SetPC(0x0300)
	for pc, _ := a.cpu.GetPCAndSP(); pc != 0x0306; pc, _ = a.cpu.GetPCAndSP() {
		a.stepInstruction()
	}
}
