package shared

import "testing"

func TestPaddlesUnplugged(t *testing.T) {
	p := NewPaddles(0, false)

	for i := range 4 {
		value, plugged := p.ReadPaddle(i)
		if plugged {
			t.Errorf("paddle %v should not be plugged", i)
		}
		if value != paddleUnplugged {
			t.Errorf("paddle %v is %v, expected %v", i, value, paddleUnplugged)
		}
	}

	for i := range 3 {
		if p.ReadButton(i) {
			t.Errorf("button %v should not be pressed", i)
		}
	}
}

func TestPaddlesAxis(t *testing.T) {
	p := NewPaddles(2, false)

	cases := []struct {
		joystick int
		axis     int
		value    int16
		paddle   int
		expected uint8
	}{
		{0, 0, -32768, 0, 0},  // Full left
		{0, 1, 0, 1, 128},     // Centered
		{1, 0, 32767, 2, 255}, // Full right
		{1, 1, -32768, 3, 0},
	}

	for _, c := range cases {
		p.SetAxis(c.joystick, c.axis, c.value)
		actual, plugged := p.ReadPaddle(c.paddle)
		if actual != c.expected || !plugged {
			t.Errorf("SetAxis(%v, %v, %v) leaves paddle %v at (%v, %v), expected (%v, true)",
				c.joystick, c.axis, c.value, c.paddle, actual, plugged, c.expected)
		}
	}

	// The joysticks and axis beyond the ones the Apple II has are ignored
	p.SetAxis(2, 0, 32767)
	p.SetAxis(0, 2, 32767)
	p.SetAxis(-1, 0, 32767)
}

func TestPaddlesButtons(t *testing.T) {
	cases := []struct {
		joystick int
		button   int
		expected int // The button of the machine it is read as
	}{
		{0, 0, 0}, // Primary of the first joystick
		{0, 1, 1}, // Secondary of the first, also primary of the second
		{1, 0, 1},
		{1, 1, 2}, // Secondary of the second
		{0, 2, 0}, // The buttons beyond the second repeat
	}

	for _, c := range cases {
		p := NewPaddles(2, false)
		p.SetButton(c.joystick, c.button, true)

		for i := range 3 {
			expected := i == c.expected
			if p.ReadButton(i) != expected {
				t.Errorf("button %v of joystick %v: button %v of the machine is %v, expected %v",
					c.button, c.joystick, i, p.ReadButton(i), expected)
			}
		}
	}
}

func TestPaddlesAppleKeys(t *testing.T) {
	p := NewPaddles(1, false)

	p.SetOpenApple(true)
	p.SetClosedApple(true)
	if !p.ReadButton(0) || !p.ReadButton(1) {
		t.Error("the apple keys should be read as the first two buttons")
	}

	p.ReleaseAppleKeys()
	if p.ReadButton(0) || p.ReadButton(1) {
		t.Error("the apple keys should be released")
	}
}

func TestPaddlesMouse(t *testing.T) {
	p := NewPaddles(0, true)

	// The mouse takes the place of the first joystick, centered
	value, plugged := p.ReadPaddle(0)
	if !plugged || value != paddleCentered {
		t.Errorf("the mouse paddle is (%v, %v), expected (%v, true)",
			value, plugged, paddleCentered)
	}

	cases := []struct {
		x, y     int
		expected [2]uint8
	}{
		{200, 100, [2]uint8{127, 127}}, // The center of the window
		{0, 0, [2]uint8{0, 27}},
		{400, 200, [2]uint8{255, 227}},
	}

	for _, c := range cases {
		p.SetMousePosition(c.x, c.y, 400, 200)
		x, _ := p.ReadPaddle(0)
		y, _ := p.ReadPaddle(1)
		if x != c.expected[0] || y != c.expected[1] {
			t.Errorf("the pointer at %v,%v leaves the paddles at %v,%v, expected %v",
				c.x, c.y, x, y, c.expected)
		}
	}

	p.SetMouseButton(MouseButtonRight, true)
	if !p.ReadButton(1) {
		t.Error("the right button should be read as the second button")
	}
}

func TestPaddlesMouseWithJoystick(t *testing.T) {
	// With a joystick plugged the mouse is left for the machine to use
	p := NewPaddles(1, true)

	p.SetAxis(0, 0, 0)
	p.SetMousePosition(0, 0, 400, 200)
	p.SetMouseButton(MouseButtonLeft, true)

	if value, _ := p.ReadPaddle(0); value != 128 {
		t.Errorf("the paddle is %v, the mouse should not have moved it", value)
	}
	if p.ReadButton(0) {
		t.Error("the mouse should not press the buttons")
	}
}
