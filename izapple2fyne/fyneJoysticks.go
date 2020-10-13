package main

import (
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

/*
  Apple 2 supports four paddles and 3 pushbuttons. The first two paddles are
the X, Y axis of the first joystick. The second two correspond the the second
joystick.
  Button 0 is the primary button of joystick 0.
  Button 1 is the secondary button of joystick 0 but also the primary button of
joystick 1.
  Button 2 is the secondary button of Joystick 1.
*/

// TODO: key as buttons as on the IIe and mouse as joystick

type joystickInfo struct {
	present bool
	name    string
	paddles [2]uint8
	buttons [2]bool
}

type joysticks struct {
	s    *state
	info [2]*joystickInfo
}

const unplugged = uint8(255) // Max resistance when unplugged

func newJoysticks(s *state) *joysticks {
	var j joysticks
	j.s = s
	return &j
}

func (j *joysticks) start() {
	pool := time.NewTicker(time.Second / 50)
	go func() {
		runtime.LockOSThread()

		for {
			select {
			case <-pool.C:
				j.info[0] = j.queryJoystick(glfw.Joystick1)
				j.info[1] = j.queryJoystick(glfw.Joystick2)

				j.s.devices.joystick.updateJoy1(j.info[0])
				j.s.devices.joystick.updateJoy2(j.info[1])
			}
		}
	}()

}

func (j *joysticks) queryJoystick(joy glfw.Joystick) *joystickInfo {
	if !joy.Present() {
		return nil
	}

	var info joystickInfo
	info.name = joy.GetName()
	buttons := joy.GetButtons()
	for i, b := range buttons {
		if b == glfw.Press {
			info.buttons[i%2] = true
		}
	}
	axes := joy.GetAxes()
	for i := 0; i < len(info.paddles); i++ {
		info.paddles[i] = unplugged
		if i < len(axes) {
			v := uint16((axes[i] + 1.0) / 2.0 * 256.0)
			if v > 255 {
				v = 255
			}
			info.paddles[i] = uint8(v)
		}
	}
	return &info
}

func (j *joysticks) ReadButton(i int) bool {
	var value bool
	i0 := j.info[0]
	i1 := j.info[1]
	switch i {
	case 0:
		value = (i0 != nil) && i0.buttons[0]
	case 1:
		// It can be secondary of first or primary of second
		value = ((i0 != nil) && i0.buttons[1]) ||
			(i1 != nil) && i1.buttons[0]
	case 2:
		value = (i1 != nil) && i1.buttons[0]
	}
	return value
}

func (j *joysticks) ReadPaddle(i int) (uint8, bool) {
	var value = unplugged
	info := j.info[i/2]
	if info != nil {
		value = info.paddles[i%2]
	}
	return value, value != unplugged
}
