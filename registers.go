package main

const (
	regA = 0
	regX = 1
	regY = 2
	// empty
	regP  = 4
	regS  = 5
	regPC = 6 // 2 bytes
)

type registers struct {
	data [8]uint8
}

func (r registers) getA() uint8 { return r.data[regA] }
func (r registers) getX() uint8 { return r.data[regX] }
func (r registers) getY() uint8 { return r.data[regY] }
func (r registers) getP() uint8 { return r.data[regP] }
func (r registers) getS() uint8 { return r.data[regS] }

func (r registers) setRegister(i int, v uint8) registers {
	r.data[i] = v
	return r
}
func (r registers) setA(v uint8) registers { return r.setRegister(regA, v) }
func (r registers) setX(v uint8) registers { return r.setRegister(regX, v) }
func (r registers) setY(v uint8) registers { return r.setRegister(regY, v) }
func (r registers) setP(v uint8) registers { return r.setRegister(regP, v) }
func (r registers) setS(v uint8) registers { return r.setRegister(regS, v) }

func (r registers) getPC() uint16 {
	return uint16(r.data[regPC])*256 + uint16(r.data[regPC+1])
}

func (r registers) setPC(v uint16) registers {
	r.data[regPC] = uint8(v >> 8)
	r.data[regPC+1] = uint8(v)
	return r
}
