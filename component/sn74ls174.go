package component

/*
	SN74LS174: hex D flip-flop

	http://www.skot9000.com/ttl/datasheets/174.pdf

Pins:
	A0, A1, A2: Address inputs. 'address' parameter of Write()
	D: Data input. 'value' parameter of Write()
	E: Enable input. 'enable' parameter of Write()
	C: Clear input. Method Reset()
	Q0, Q7: Parallel latch outputs. Return of Q(n)
*/

type SN74LS174 struct {
	q [6]bool
}

func (o *SN74LS174) Update(d0, d1, d2, d3, d4, d5, mr bool) {
	o.q[0] = d0 && mr
	o.q[1] = d1 && mr
	o.q[2] = d2 && mr
	o.q[3] = d3 && mr
	o.q[4] = d4 && mr
	o.q[5] = d5 && mr
}

func (o *SN74LS174) Reset() {
	o.Update(false, false, false, false, false, false, false)
}

func (o *SN74LS174) Q(index uint8) bool {
	if index > 5 {
		panic("There are only 6 flip flops")
	}
	return o.q[index]
}
