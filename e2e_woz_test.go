package izapple2

import (
	"testing"
)

func testWoz(t *testing.T, sequencer bool, file string, expectedTracks []int, cycleLimit uint64) {

	overrides := newConfiguration()
	if sequencer {
		overrides.set(confS6, "diskiiseq,disk1=\"woz_test_images/"+file+"\"")
	} else {
		overrides.set(confS6, "diskii,disk1=\"woz_test_images/"+file+"\"")
	}
	at, err := makeApple2Tester("2enh", overrides)
	if err != nil {
		t.Fatal(err)
	}

	diskIIcard, ok := at.a.cards[6].(cardDisk2Shared)
	if !ok {
		t.Fatal("Not a disk II card")
	}
	tt := makeTrackTracerSummary()
	diskIIcard.setTrackTracer(tt)

	expectedLen := len(expectedTracks)

	at.terminateCondition = func(a *Apple2) bool {
		tracksMayMatch := len(tt.quarterTracks) >= expectedLen &&
			tt.quarterTracks[expectedLen-1] == expectedTracks[expectedLen-1]

		return tracksMayMatch || a.cpu.GetCycles() > cycleLimit
	}
	at.run()

	if !tt.isTraceAsExpected(expectedTracks) {
		t.Errorf("Quarter tracks, expected %#v, got %#v", expectedTracks, tt.quarterTracks)
	}

	// t.Errorf("Cycles: %d vs  %d", at.a.cpu.GetCycles(), cycleLimit)
}

const (
	all  = 0
	seq  = 1 // Passes only with the sequencer implementation
	none = 2 // Fails also with the sequencer implementation
)

func TestWoz(t *testing.T) {
	testCases := []struct {
		name           string
		skip           int
		disk           string
		cycleLimit     uint64
		expectedTracks []int
	}{
		// How to begin
		{"DOS 3.2", all, "DOS 3.2 System Master.woz", 7_000_000, []int{0, 72}},
		{"DOS 3.3", all, "DOS 3.3 System Master.woz", 11_000_000, []int{0, 8, 0, 76, 68, 84, 68, 84, 68, 92, 16, 24}},

		// Next choices
		{"Bouncing Kamungas", all, "Bouncing Kamungas - Disk 1, Side A.woz", 30_000_000, []int{0, 32, 0, 40, 0}},
		// Runs but the test is unstable {"Commando", seq, "Commando - Disk 1, Side A.woz", 15_000_000, []int{0, 136, 68, 128, 68, 128, 68, 124, 12, 116, 108}},
		{"Planetfall", all, "Planetfall - Disk 1, Side A.woz", 4_000_000, []int{0, 8}},
		{"Rescue Raiders", all, "Rescue Raiders - Disk 1, Side B.woz", 80_000_000, []int{
			0, 84, 44, 46,
			0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0, 2, 0,
			84, 44, 116, 4, 8, 4, 12, 8, 84, 44, 132, 0, 120, 44, 84, 44, 124, 0, 120, 44}},
		{"Sammy Lightfoot", all, "Sammy Lightfoot - Disk 1, Side A.woz", 80_000_000, []int{0, 64, 8, 20}},
		{"Stargate", all, "Stargate - Disk 1, Side A.woz", 50_000_000, []int{0, 8, 0, 72, 68, 80, 68, 80, 12, 44}},

		// Cross track sync
		{"Blazing Paddles", all, "Blazing Paddles (Baudville).woz", 6_000_000, []int{0, 28, 0, 16, 12, 56, 52, 80}},
		{"Take 1", all, "Take 1 (Baudville).woz", 8_000_000, []int{0, 28, 0, 4, 0, 72, 0, 20, 0}},
		{"Hard Hat Mack", all, "Hard Hat Mack - Disk 1, Side A.woz", 10_000_000, []int{0, 134, 132}},

		// Half tracks
		{"The Bilestoad", all, "The Bilestoad - Disk 1, Side A.woz", 6_000_000, []int{0, 24}},

		// Even more bit fiddling
		{"Dino Eggs", all, "Dino Eggs - Disk 1, Side A.woz", 9_000_000, []int{0, 78, 60, 108, 32}},
		{"Crisis Mountain", all, "Crisis Mountain - Disk 1, Side A.woz", 20_000_000, []int{0, 32, 8, 32, 20, 76, 20, 36, 32, 84, 52, 64}},
		{"Miner 2049er II", all, "Miner 2049er II - Disk 1, Side A.woz", 11_000_000, []int{0, 12, 8, 32, 12, 136, 132}},

		// When bits aren't really bits
		{"The Print Shop Companion", all, "The Print Shop Companion - Disk 1, Side A.woz", 14_000_000, []int{0, 68, 44, 68, 40, 68, 40, 136, 60}},

		// What is the lifepsan of the data latch?
		{"First Math Adventures", seq, "First Math Adventures - Understanding Word Problems.woz", 6_000_000, []int{0, 8, 0, 68, 12, 20}},

		// Reading Offset Data Streams
		{"Wings of Fury", seq, "Wings of Fury - Disk 1, Side A.woz", 410_000_000, []int{0, 4, 0, 136, 124, 128, 24, 136, 124, 128, 24, 136, 124, 128, 24, 136, 124, 128, 24, 104}},
		{"Stickybear Town Builder", all, "Stickybear Town Builder - Disk 1, Side A.woz", 8_000_000, []int{0, 16, 12, 112, 80, 100, 8}},

		// Optimal bit timing
		// Requires disk change {"Border Zone", "Border Zone - Disk 1, Side A.woz", 500_000_000, []int{1,1,1,1,1,1}},

		// Extra
		{"Mr. Do", seq, "Mr. Do.woz", 95_000_000, []int{0, 108, 48, 104, 72, 84, 0, 4}},
		{"Wavy Navy", all, "Wavy Navy.woz", 9_000_000, []int{0, 136}},
		// SAGA6 requires disk change,
		// Note that Congo Bongo works with the non sequencer implementation but the test is unstable
		{"Congo Bongo", seq, "Congo Bongo.woz", 8_000_000, []int{0, 4, 2, 40, 20, 40, 16, 124, 116}},
		// Wizardry III requires disk change,

	}

	for _, tc := range testCases {
		if tc.skip == all {
			t.Run(tc.name, func(t *testing.T) {
				testWoz(t, false, tc.disk, tc.expectedTracks, tc.cycleLimit)
			})
		}
		if tc.skip == all || tc.skip == seq {
			t.Run(tc.name+" SEQ", func(t *testing.T) {
				testWoz(t, true, tc.disk, tc.expectedTracks, tc.cycleLimit)
			})
		}

	}
}
