package screen

import (
	"testing"
)

func TestTextMemoryByteToString(t *testing.T) {
	charExpectation(t, 0x01, false, "\033[7mA\033[0m")
	charExpectation(t, 0x21, false, "\033[7m!\033[0m")
	charExpectation(t, 0x41, false, "\033[5mA\033[0m")
	charExpectation(t, 0x61, false, "\033[5m!\033[0m")
	charExpectation(t, 0x81, false, "A")
	charExpectation(t, 0xa1, false, "!")
	charExpectation(t, 0xc1, false, "A")
	charExpectation(t, 0xe1, false, "!")

	charExpectation(t, 0x01, true, "\033[7mA\033[0m")
	charExpectation(t, 0x21, true, "\033[7m!\033[0m")
	charExpectation(t, 0x41, true, "\033[7mA\033[0m")
	charExpectation(t, 0x61, true, "\033[7ma\033[0m")
	charExpectation(t, 0x81, true, "A")
	charExpectation(t, 0xa1, true, "!")
	charExpectation(t, 0xc1, true, "A")
	charExpectation(t, 0xe1, true, "a")

}

func charExpectation(t *testing.T, arg uint8, alt bool, expect string) {
	s := textMemoryByteToString(arg, alt, alt)
	if s != expect {
		t.Errorf("For 0x%02x:%v, got %v, expected %v", arg, alt, s, expect)
	}
}
