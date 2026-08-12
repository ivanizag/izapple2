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

	m.x = mousePosition(x, width)
	m.y = mousePosition(y, height)
}

// SetButton presses or releases the button of the mouse
func (m *Mouse) SetButton(pressed bool) {
	m.pressed = pressed
}

// mousePosition turns a position on the window into the full range the machine
// reads the mouse in, keeping what is outside the window on the edge
func mousePosition(position int, size int) uint16 {
	return uint16(max(min(65536*position/size, 65535), 0))
}

// izapple2.MouseProvider implementation

// ReadMouse returns the position of the pointer on the window and whether its
// button is pressed
func (m *Mouse) ReadMouse() (x uint16, y uint16, pressed bool) {
	return m.x, m.y, m.pressed
}
