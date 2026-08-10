package shared

// Mouse is the position and the button of the pointer, as the mouse card of
// the machine reads them
type Mouse struct {
	x       uint16
	y       uint16
	pressed bool
}

func NewMouse() *Mouse {
	var m Mouse
	return &m
}

// SetPosition places the pointer, from where it is on a window of the given
// size
func (m *Mouse) SetPosition(x int, y int, width int, height int) {
	if width <= 0 || height <= 0 {
		// The window is not sized yet
		return
	}

	m.x = uint16(65536 * x / width)
	m.y = uint16(65536 * y / height)
}

// SetButton presses or releases the button of the mouse
func (m *Mouse) SetButton(pressed bool) {
	m.pressed = pressed
}

// ReadMouse returns the position of the pointer on the window and whether its
// button is pressed
func (m *Mouse) ReadMouse() (x uint16, y uint16, pressed bool) {
	return m.x, m.y, m.pressed
}

// TODO: SDL_WarpMouseInWindow
