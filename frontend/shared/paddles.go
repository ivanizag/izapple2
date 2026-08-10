package shared

/*
The Apple II supports four paddles and 3 pushbuttons. The first two paddles are
the X, Y axis of the first joystick. The second two correspond to the second
joystick.

	Button 0 is the primary button of joystick 0.
	Button 1 is the secondary button of joystick 0 but also the primary button
	of joystick 1.
	Button 2 is the secondary button of joystick 1.

The Apple //e reads the open and closed apple keys as the first two buttons.
Where there is no joystick, the mouse can move the first two paddles.
*/

// The buttons of the mouse, when it is what moves the paddles
const (
	MouseButtonLeft = iota
	MouseButtonRight
	MouseButtonMiddle
)

const (
	paddleUnplugged = 255 // Max resistance when unplugged
	paddleCentered  = 127
)

// Paddles builds the paddles and buttons the machine reads from the events of
// a frontend
type Paddles struct {
	paddle       [4]uint8
	hasPaddle    [4]bool
	button       [4]bool
	appleKeys    [2]bool // The open and the closed apple keys
	mouseButtons [3]bool
	useMouse     bool
}

/*
NewPaddles creates the paddles of a machine with the given number of joysticks,
of which only the first two are used.

Without joysticks, useMouseAlt moves the first two paddles with the mouse.
*/
func NewPaddles(joysticks int, useMouseAlt bool) *Paddles {
	var p Paddles

	for i := range p.paddle {
		p.paddle[i] = paddleUnplugged
	}
	for slot := 0; slot < joysticks && slot < 2; slot++ {
		p.hasPaddle[slot*2] = true
		p.hasPaddle[slot*2+1] = true
	}

	if useMouseAlt && !p.hasPaddle[0] {
		// Use the mouse as joystick
		p.useMouse = true
		p.hasPaddle[0] = true
		p.hasPaddle[1] = true
		p.paddle[0] = paddleCentered
		p.paddle[1] = paddleCentered
	}

	// To enter Apple IIe on self test mode
	// p.appleKeys[1] = true

	return &p
}

// SetAxis moves a paddle of a joystick, with the value on the full range of
// the axis
func (p *Paddles) SetAxis(joystick int, axis int, value int16) {
	if joystick < 0 || joystick >= 2 || axis < 0 || axis >= 2 {
		// Only the first two axis of the first two joysticks
		return
	}

	p.paddle[joystick*2+axis] = uint8((value >> 8) + 128)
}

// SetButton presses or releases a button of a joystick. The buttons beyond the
// second are read as the second one.
func (p *Paddles) SetButton(joystick int, button int, pressed bool) {
	if joystick < 0 || joystick >= 2 || button < 0 {
		// Only the buttons of the first two joysticks
		return
	}

	p.button[joystick*2+(button%2)] = pressed
}

// SetOpenApple presses or releases the open apple key, the first button
func (p *Paddles) SetOpenApple(pressed bool) {
	p.appleKeys[0] = pressed
}

// SetClosedApple presses or releases the closed apple key, the second button
func (p *Paddles) SetClosedApple(pressed bool) {
	p.appleKeys[1] = pressed
}

// ReleaseAppleKeys forgets the apple keys being pressed. The key up event of a
// key held while the window loses the focus goes to whoever took it, leaving
// the apple key pressed forever.
func (p *Paddles) ReleaseAppleKeys() {
	p.appleKeys = [2]bool{}
}

// SetMousePosition moves the paddles with the pointer on a window of the given
// size, where the mouse is what the machine has instead of a joystick. The
// paddles are centered on the center of the window.
func (p *Paddles) SetMousePosition(x int, y int, width int, height int) {
	if !p.useMouse {
		return
	}

	p.paddle[0] = mouseToPaddle(x, width)
	p.paddle[1] = mouseToPaddle(y, height)
}

// SetMouseButton presses or releases a button of the mouse, where the mouse is
// what the machine has instead of a joystick
func (p *Paddles) SetMouseButton(button int, pressed bool) {
	if !p.useMouse || button < 0 || button >= len(p.mouseButtons) {
		return
	}

	p.mouseButtons[button] = pressed
}

// ReadButton returns whether a pushbutton is pressed, as the machine reads it
func (p *Paddles) ReadButton(i int) bool {
	switch i {
	case 0:
		return p.button[0] || p.appleKeys[0] || p.mouseButtons[MouseButtonLeft]
	case 1:
		// It can be secondary of first or primary of second
		return p.button[1] || p.button[2] || p.appleKeys[1] ||
			p.mouseButtons[MouseButtonRight]
	case 2:
		return p.button[3] || p.mouseButtons[MouseButtonMiddle]
	}

	return false
}

// ReadPaddle returns the resistance of a paddle and whether it is plugged
func (p *Paddles) ReadPaddle(i int) (uint8, bool) {
	return p.paddle[i], p.hasPaddle[i]
}

// mouseToPaddle turns a position on the window into the value of a paddle
// centered on the center of the window
func mouseToPaddle(x int, width int) uint8 {
	return uint8(max(min(x-(width/2)+paddleCentered, 255), 0))
}
