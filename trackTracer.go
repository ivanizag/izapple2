package izapple2

type trackTracer interface {
	traceTrack(quarterTrack int)
}

type trackTracerSummary struct {
	quarterTracks []int
}

func makeTrackTracerSummary() *trackTracerSummary {
	var tt trackTracerSummary
	tt.quarterTracks = make([]int, 0, 100)
	return &tt
}

func (tt *trackTracerSummary) traceTrack(quarterTrack int) {
	if tt == nil {
		return
	}

	length := len(tt.quarterTracks)
	if length == 0 {
		// Second change, just record
		tt.quarterTracks = append(tt.quarterTracks, quarterTrack)
		return
	}

	last := tt.quarterTracks[length-1]
	if quarterTrack == last {
		// No changes
		return
	}

	if length == 1 {
		// Second change, just record
		tt.quarterTracks = append(tt.quarterTracks, quarterTrack)
		return
	}

	// We don't want to registers the initial jumps around 0 seen when initializing the disk to track 0
	prevToLast := tt.quarterTracks[length-2]
	if length == 2 && prevToLast == 0 && (last == 1 || last == 2) && quarterTrack == 0 {
		tt.quarterTracks = tt.quarterTracks[0:0]
	}

	// We don't want to track each increment. If tracks goes from 1 to 14, we just want 1 and 14.
	wasGoingUp := last > prevToLast
	isGoingUp := quarterTrack > last
	if isGoingUp == wasGoingUp {
		// Same direction, update the last registry
		tt.quarterTracks[length-1] = quarterTrack
	} else {
		// Change direction, add a new registry
		tt.quarterTracks = append(tt.quarterTracks, quarterTrack)
	}
}

func (tt *trackTracerSummary) isTraceAsExpected(expected []int) bool {
	if len(tt.quarterTracks) != len(expected) {
		return false
	}

	for i, v := range tt.quarterTracks {
		if v != expected[i] {
			return false
		}
	}

	return true
}
