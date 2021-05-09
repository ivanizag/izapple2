package component

func ByteToPins(v uint8) [8]bool {
	var pins [8]bool
	for i := 0; i < 8; i++ {
		pins[i] = (v & 1) != 0
		v >>= 1
	}
	return pins
}

func PinsToByte(pins [8]bool) uint8 {
	v := uint8(0)
	for i := 7; i >= 0; i-- {
		v <<= 1
		if pins[i] {
			v++
		}
	}
	return v
}

func ReversePins(data uint8) uint8 {
	pins := ByteToPins(data)
	return PinsToByte([8]bool{
		pins[7],
		pins[6],
		pins[5],
		pins[4],
		pins[3],
		pins[2],
		pins[1],
		pins[0],
	})
}
