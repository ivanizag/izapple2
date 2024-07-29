package izapple2

/*
Stepper motor to position the track.

There are a number of group of four magnets. The stepper motor can be thought as a long
line of groups of magnets, each group on the same configuration. We call phase each of those
magnets. The cog is attracted to the enabled magnets, and can stay aligned to a magnet or
between two.

Phases (magnets):                       3   2   1   0   3   2   1   0   3   2   1   0
Cog direction (step within a group):    7 6 5 4 3 2 1 0 7 6 5 4 3 2 1 0 7 6 5 4 3 2 1 0

We will consider that the cog would go to the prefferred position if there is one. Independently
of the previous position. The previous position is only used to know if it goes up or down
a full group.

Phases are coded in 4 bits in an uint8. Q3, q2, q1 and q0 with q0 on the LSB.
*/

const (
	undefinedPosition = -1
	maxStep           = 68 * 2 // What is the maximum quarter tracks a DiskII can go?
	stepsPerGroup     = 8
	stepsPerTrack     = 4
)

var cogPositions = []int{
	undefinedPosition, // 0000, phases active
	0,                 // 0001
	2,                 // 0010
	1,                 // 0011
	4,                 // 0100
	undefinedPosition, // 0101
	3,                 // 0110
	2,                 // 0111
	6,                 // 1000
	7,                 // 1001
	undefinedPosition, // 1010
	0,                 // 1011
	5,                 // 1100
	6,                 // 1101
	4,                 // 1110
	undefinedPosition, // 1111
}

func moveDriveStepper(phases uint8, prevStep int) int {

	// fmt.Printf("magnets: 0x%x\n", phases)

	cogPosition := cogPositions[phases]
	if cogPosition == undefinedPosition {
		// Don't move if magnets don't push on a defined direction.
		return prevStep
	}

	prevPosition := prevStep % stepsPerGroup // Direction, step in the current group of magnets.
	delta := cogPosition - prevPosition
	if delta < 0 {
		delta += stepsPerGroup
	}

	var nextStep int
	if delta < 4 {
		// Steps up
		nextStep = prevStep + delta
		if nextStep > maxStep {
			nextStep = maxStep
		}
	} else if delta == 4 {
		// Don't move if magnets push on the opposite direction
		nextStep = prevStep
	} else { // delta > 4
		// Steps down
		nextStep = prevStep + delta - stepsPerGroup
		if nextStep < 0 {
			nextStep = 0
		}
	}

	// fmt.Printf("[DiskII] 1/4 track: %03d %vO\n", nextStep, strings.Repeat(" ", nextStep))
	return nextStep
}
