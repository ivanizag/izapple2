package izapple2

import (
	"bytes"
	"testing"
)

func TestSmartPortPacketRoundTrip(t *testing.T) {
	for _, length := range append([]int{512}, makeRange(31)...) {
		data := make([]uint8, length)
		for i := range data {
			data[i] = uint8(i*37 + length)
		}
		packet := append([]uint8{0xc3}, smartPortEncodePacket(0x81, 0x80, smartPortBusPacketData, 0x27, data)...)
		if len(packet) != smartPortPacketLength(uint8(length%7), uint8(length/7)) {
			t.Errorf("length %v: unexpected packet length %v", length, len(packet))
		}
		destination, packetType, decoded, ok := smartPortDecodePacket(packet)
		if !ok || destination != 0x01 || packetType != smartPortBusPacketData || !bytes.Equal(decoded, data) {
			t.Errorf("length %v: got %v %x %x %v", length, destination, packetType, decoded, ok)
		}
		for _, value := range packet {
			if value&0x80 == 0 {
				t.Errorf("length %v: a byte without the high bit, $%02x", length, value)
				break
			}
		}
	}
}

func makeRange(n int) []int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	return values
}

func TestSmartPortPacketErrors(t *testing.T) {
	packet := append([]uint8{0xc3}, smartPortEncodePacket(0x81, 0x80, smartPortBusPacketCommand, 0, []uint8{1, 3, 0, 8, 0, 0, 0, 0, 0})...)

	bad := append([]uint8{}, packet...)
	bad[9] ^= 0x01 // A data byte
	if _, _, _, ok := smartPortDecodePacket(bad); ok {
		t.Error("expected the checksum error")
	}
	bad = append([]uint8{}, packet...)
	bad[len(bad)-1] = 0xff
	if _, _, _, ok := smartPortDecodePacket(bad); ok {
		t.Error("expected the end mark error")
	}
}

// testSmartPortDevice answers with the block number on the data
type testSmartPortDevice struct {
	calls []*smartPortCall
}

func (d *testSmartPortDevice) exec(call *smartPortCall) (uint8, []uint8) {
	d.calls = append(d.calls, call)
	return smartPortNoError, []uint8{call.param8(4), 0xaa}
}

func TestSmartPortBusExchange(t *testing.T) {
	device := &testSmartPortDevice{}
	bus := newSmartPortBus()
	bus.addDevice(device)
	const enabled = 0x0a // PHASE1 and PHASE3
	const req = 0x01     // PHASE0

	send := func(packet []uint8) {
		bus.SetLines(enabled|req, false, true)
		for _, value := range append([]uint8{0xff, 0x3f, 0xcf, 0xf3, 0xfc, 0xff, 0xc3}, packet...) {
			bus.Write(value)
		}
		if bus.Sense() {
			t.Fatal("expected the ACK of the packet")
		}
		bus.SetLines(enabled, false, true)
	}
	receive := func() []uint8 {
		if !bus.Sense() {
			t.Fatal("expected the reply ready")
		}
		bus.SetLines(enabled|req, false, true)
		var reply []uint8
		for {
			value, ok := bus.Read()
			if !ok {
				break
			}
			reply = append(reply, value)
		}
		if bus.Sense() {
			t.Fatal("expected the ACK after the reply")
		}
		bus.SetLines(enabled, false, true)
		start := bytes.IndexByte(reply, 0xc3)
		_, _, data, ok := smartPortDecodePacket(reply[start:])
		if !ok {
			t.Fatalf("bad reply % x", reply)
		}
		return data
	}

	// INIT gives the unit 1, the only device is the last one
	send(smartPortEncodePacket(0x81, 0x80, smartPortBusPacketCommand, 0, []uint8{smartPortCommandInit, 2, 0, 8, 0, 0, 0, 0, 0}))
	receive()
	if bus.ids[0] != 1 {
		t.Errorf("expected the unit 1, got %v", bus.ids[0])
	}

	// A command for a unit that is not there is not taken
	bus.SetLines(enabled|req, false, true)
	for _, value := range append([]uint8{0xc3}, smartPortEncodePacket(0x82, 0x80, smartPortBusPacketCommand, 0, []uint8{1, 3, 0, 8, 7, 0, 0, 0, 0})...) {
		bus.Write(value)
	}
	if !bus.Sense() {
		t.Error("expected no ACK for the unit 2")
	}
	bus.SetLines(enabled, false, true)

	// A READBLOCK of the block 7
	send(smartPortEncodePacket(0x81, 0x80, smartPortBusPacketCommand, 0, []uint8{smartPortCommandReadBlock, 3, 0, 8, 7, 0, 0, 0, 0}))
	if data := receive(); !bytes.Equal(data, []uint8{7, 0xaa}) {
		t.Errorf("unexpected reply % x", data)
	}
	call := device.calls[len(device.calls)-1]
	if call.command != smartPortCommandReadBlock || call.unit() != 1 || call.param24(4) != 7 {
		t.Errorf("unexpected call %v", call)
	}

	// After a reset the units have no number
	bus.SetLines(0x05, false, false)
	if bus.ids[0] != 0 {
		t.Error("expected the units reset")
	}
}
