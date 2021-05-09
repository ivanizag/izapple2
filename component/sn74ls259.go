package component

/*
	SN74LS259 or 9344: 8-bit addressable latch

	http://www.skot9000.com/ttl/datasheets/259.pdf

Pins:
	A0, A1, A2: Address inputs. 'address' parameter of Write()
	D: Data input. 'value' parameter of Write()
	E: Enable input. 'enable' parameter of Write()
	C: Clear input. Method Reset()
	Q0, Q7: Parallel latch outputs. Return of Q(n)
*/

type SN74LS259 struct {
	value uint8
}

func (o *SN74LS259) Write(address uint8, value bool, enable bool) {
	if address > 7 {
		panic("The address can have only three bits")
	}

	if !enable {
		return
	}

	o.value = o.value &^ (uint8(1) << address)
	if value {
		o.value |= 1 << address
	}
}

func (o *SN74LS259) Q(address uint8) bool {
	if address > 7 {
		panic("The address can have only three bits")
	}

	return ((o.value >> address) & 1) == 1
}

func (o *SN74LS259) Reset() {
	o.value = 0
}
