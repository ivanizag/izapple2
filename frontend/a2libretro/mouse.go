package main

/*
mouse reports a mouse that never moves and is never pressed.

The core does not map the host pointer to the Apple II mouse, but the mouse
card reads its provider without checking that there is one, so the models that
carry the card, like "desktop", need one to be attached.
*/
type mouse struct{}

func newMouse() *mouse {
	return &mouse{}
}

func (m *mouse) ReadMouse() (x uint16, y uint16, pressed bool) {
	return 0, 0, false
}
