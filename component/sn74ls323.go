package component

/*
	SN74LS323: 8-bit shift storage register

	http://www.skot9000.com/ttl/datasheets/323.pdf

Pins:
	CP: Clock pulse. Call to Update()
	DS0: Serial data input for shift right. 'ds0' parameter of Update()
	DS7: Serial data input for shift left. 'ds1' parameter of Update()
	I0-7: Parallel data input. 'data' parameter of Update()
	O0-7: Parallel data output. Return of Output()
	OE1-2: Not supported
	Q0: Serial outputs LSB. Return of Q0()
	Q7: Serial outputs MSB. Return of Q7()
	S0: Mode select, 's0' parameter of Update()
	S1: Mode select, 's1' parameter of Update()
	SR: Sync reset (active low), 'sr_n' parameter of Update()

Note: left for datasheet (I0-I1-I2..I7) is right for uint8 (b7-b6...b0)
*/

type SN74LS323 struct {
	value uint8
}

func (o *SN74LS323) Update(data uint8, s0 bool, s1 bool, sr_n bool, ds0 bool, ds7 bool) {
	if !sr_n {
		// Reset on SR low
		o.value = 0
		return
	}

	if s1 {
		if s0 {
			// high, high: Parallel load
			o.value = data
		} else {
			// high, low: Shift left (it's shift right for uint8)
			o.value >>= 1
			if ds7 {
				o.value += 0x80
			}
		}
	} else {
		if s0 {
			// low, high: Shift right (it's shift left for uint8)
			o.value <<= 1
			if ds0 {
				o.value += 0x01
			}
		} else {
			// low, low: Hold
			// do nothing
		}
	}
}

func (o *SN74LS323) Q0() bool {
	return (o.value & 1) != 0
}

func (o *SN74LS323) Q7() bool {
	return (o.value & 0x80) != 0
}

func (o *SN74LS323) Output() uint8 {
	return o.value
}
