package izapple2

import (
	"fmt"
)

/*
smartPortBus is the SmartPort bus with its devices, as connected to the
drive port of an IWM. It plays the devices side of the protocol: it takes
the command packets the host writes, runs them on the devices and gives the
reply packets for the host to read.

See:

	The firmware of the Liron card and of the UniDisk 3.5, disassembled by
	Eric Smith: https://git.applefritter.com/Apple-2-SW/a2-liron

The lines are the phases of the IWM of the host, and the SENSE input:

	PHASE1 and PHASE3: bus enabled
	PHASE0: REQ, the host is sending or wants to receive a packet
	PHASE0 and PHASE2 without the bus enabled: reset of the bus
	SENSE: ACK, low when the device has taken or sent a packet

A packet is a sync sequence ending in $C3, then seven header bytes:
destination, source, type, aux type, status, number of odd bytes and number
of groups. The data follows: the first bytes, length modulo 7, as the odd
bytes, the rest in groups of seven. Each part starts with a byte with the
high bits of its bytes, all are sent with the high bit set. Then two bytes
of checksum and the end mark $C8.

A command packet, type $80, has nine bytes: the command, the parameter
count and the parameters after the unit number. WRITEBLOCK, CONTROL and
WRITE send a second packet with their data, type $82. The reply has the
result on the status byte of the header, and the data: the status list, the
block read.

INIT gives the unit numbers along the chain: after a reset the host sends
INIT to the units 1, 2... until a device answers that it is the last one.
The devices without number take them in order.
*/
type smartPortBus struct {
	devices []smartPortDevice
	ids     []uint8 // Unit number given by INIT to each device, 0 for none
	trace   bool

	state    int
	phases   uint8
	sense    bool
	received []uint8 // The packet being received, after the sync
	command  *smartPortCall
	device   int     // Index of the device of the command
	reply    []uint8 // The packet being sent
}

const (
	smartPortBusIdle          = iota
	smartPortBusReceiving     // REQ from the host, taking the packet
	smartPortBusTaken         // ACK given, waiting for the end of REQ
	smartPortBusWaitingData   // A command waiting for its data packet
	smartPortBusReplyReady    // The reply waits for REQ from the host
	smartPortBusSending       // Giving the reply packet
	smartPortBusSent          // ACK given, waiting for the end of REQ
	smartPortBusPacketCommand = 0x80
	smartPortBusPacketStatus  = 0x81
	smartPortBusPacketData    = 0x82

	smartPortBusHost    = 0x80 // Source and destination of the host
	smartPortBusEndMark = 0xc8
	smartPortBusLast    = 0x40 // INIT status of the last device
)

// The sync sequence the devices send, ending in $C3
var smartPortBusSync = []uint8{0xff, 0x3f, 0xcf, 0xf3, 0xfc, 0xff, 0xc3}

func newSmartPortBus() *smartPortBus {
	var b smartPortBus
	b.sense = true
	return &b
}

func (b *smartPortBus) addDevice(device smartPortDevice) {
	b.devices = append(b.devices, device)
	b.ids = append(b.ids, 0)
}

// component.IWMDrive implementation

// SetLines follows the phases, the enables are not used by the bus
func (b *smartPortBus) SetLines(phases uint8, _ bool, _ bool) {
	previous := b.phases
	b.phases = phases
	enabled := phases&0x0a == 0x0a
	req := phases&0x01 != 0
	wasReq := previous&0x01 != 0

	if phases&0x05 == 0x05 && phases&0x0a == 0 {
		b.reset()
		return
	}

	if req && !wasReq && enabled {
		b.requestStarted()
	} else if !req && wasReq {
		b.requestEnded()
	}
}

func (b *smartPortBus) Write(value uint8) {
	if b.state != smartPortBusReceiving {
		return
	}
	if len(b.received) == 0 && value != 0xc3 {
		// Skip the sync
		return
	}
	b.received = append(b.received, value)
	if b.isPacketComplete() {
		b.packetReceived()
	}
}

func (b *smartPortBus) Read() (uint8, bool) {
	if b.state != smartPortBusSending || len(b.reply) == 0 {
		return 0, false
	}
	value := b.reply[0]
	b.reply = b.reply[1:]
	if len(b.reply) == 0 {
		// Sent, ACK
		b.sense = false
		b.state = smartPortBusSent
	}
	return value, true
}

func (b *smartPortBus) Sense() bool {
	return b.sense
}

func (b *smartPortBus) reset() {
	for i := range b.ids {
		b.ids[i] = 0
	}
	b.state = smartPortBusIdle
	b.sense = true
	if b.trace {
		fmt.Printf("[SmartPortBus] Reset\n")
	}
}

func (b *smartPortBus) requestStarted() {
	switch b.state {
	case smartPortBusIdle, smartPortBusWaitingData:
		b.received = b.received[:0]
		b.state = smartPortBusReceiving
	case smartPortBusReplyReady:
		b.state = smartPortBusSending
	}
}

func (b *smartPortBus) requestEnded() {
	switch b.state {
	case smartPortBusTaken:
		b.sense = true
		if b.command == nil {
			b.state = smartPortBusIdle
		} else if b.command.command == smartPortCommandWriteBlock ||
			b.command.command == smartPortCommandControl ||
			b.command.command == smartPortCommandWrite {
			if b.command.data == nil {
				b.state = smartPortBusWaitingData
				return
			}
			b.execute()
		} else {
			b.execute()
		}
	case smartPortBusSent:
		b.sense = true
		b.state = smartPortBusIdle
	default:
		// A packet not taken, or not read to the end
		b.sense = true
		if b.state != smartPortBusWaitingData {
			b.state = smartPortBusIdle
		}
	}
}

// isPacketComplete returns whether the bytes received, after the sync, make
// a whole packet: the header, the data, the checksum and the end mark
func (b *smartPortBus) isPacketComplete() bool {
	if len(b.received) < 8 {
		return false
	}
	return len(b.received) >= smartPortPacketLength(b.received[6]&0x7f, b.received[7]&0x7f)
}

// smartPortPacketLength is the length of a packet after the sync byte, for
// a number of odd bytes and groups
func smartPortPacketLength(odd uint8, groups uint8) int {
	length := 1 + 7 + int(groups)*8 + 2 + 1
	if odd > 0 {
		length += 1 + int(odd)
	}
	return length
}

// packetReceived decodes the packet, and takes it with an ACK if it is for
// one of the devices
func (b *smartPortBus) packetReceived() {
	destination, packetType, data, ok := smartPortDecodePacket(b.received)
	b.state = smartPortBusIdle
	if !ok {
		if b.trace {
			fmt.Printf("[SmartPortBus] Bad packet % x\n", b.received)
		}
		return
	}

	if packetType == smartPortBusPacketData {
		if b.command == nil || b.command.data != nil {
			return
		}
		b.command.data = data
		b.taken()
		return
	}
	if packetType != smartPortBusPacketCommand || len(data) < 2 {
		return
	}

	device := -1
	if data[0] == smartPortCommandInit {
		// The first device without number takes it
		for i, id := range b.ids {
			if id == 0 {
				device = i
				break
			}
		}
	} else {
		for i, id := range b.ids {
			if id == destination {
				device = i
				break
			}
		}
	}
	if device < 0 {
		// Nobody answers
		b.command = nil
		return
	}

	// The parameters as on the SmartPort calls: the count, the unit and the
	// rest after the command
	params := append([]uint8{data[1], destination}, data[2:]...)
	b.command = newSmartPortCall(data[0], params, nil)
	b.device = device
	if data[0] == smartPortCommandInit {
		b.ids[device] = destination
	}
	b.taken()
}

func (b *smartPortBus) taken() {
	b.sense = false
	b.state = smartPortBusTaken
}

// execute runs the command on its device and prepares the reply
func (b *smartPortBus) execute() {
	call := b.command
	b.command = nil

	var result uint8
	var data []uint8
	packetType := uint8(smartPortBusPacketStatus)
	if call.command == smartPortCommandInit {
		result = smartPortNoError
		if b.isLastNumbered(b.device) {
			result = smartPortBusLast
		}
	} else {
		result, data = b.devices[b.device].exec(call)
		if call.command == smartPortCommandReadBlock || call.command == smartPortCommandRead {
			packetType = smartPortBusPacketData
		}
	}
	if b.trace {
		fmt.Printf("[SmartPortBus] %v on unit %v => %s\n", call, b.ids[b.device], smartPortErrorMessage(result))
	}

	b.reply = append(append([]uint8{}, smartPortBusSync...),
		smartPortEncodePacket(smartPortBusHost, b.ids[b.device], packetType, result, data)...)
	b.state = smartPortBusReplyReady
	b.sense = true
}

// isLastNumbered returns whether no device after this one is waiting for a
// number
func (b *smartPortBus) isLastNumbered(device int) bool {
	for _, id := range b.ids[device+1:] {
		if id == 0 {
			return false
		}
	}
	return true
}

/*
smartPortEncodePacket builds a packet, after the sync and its $C3, as the Liron firmware
sends it.
*/
func smartPortEncodePacket(destination uint8, source uint8, packetType uint8, status uint8, data []uint8) []uint8 {
	odd := len(data) % 7
	groups := len(data) / 7
	header := []uint8{
		destination | 0x80, source | 0x80, packetType | 0x80, 0x80, status | 0x80,
		uint8(odd) | 0x80, uint8(groups) | 0x80,
	}

	checksum := uint8(0)
	for _, value := range data {
		checksum ^= value
	}
	for _, value := range header {
		checksum ^= value
	}

	packet := header
	if odd > 0 {
		packet = append(packet, smartPortEncodeHighBits(data[:odd]))
		for _, value := range data[:odd] {
			packet = append(packet, value|0x80)
		}
	}
	for g := range groups {
		group := data[odd+g*7 : odd+g*7+7]
		packet = append(packet, smartPortEncodeHighBits(group))
		for _, value := range group {
			packet = append(packet, value|0x80)
		}
	}
	return append(packet, checksum|0xaa, checksum>>1|0xaa, smartPortBusEndMark)
}

// smartPortEncodeHighBits is the byte with the high bits of up to seven
// bytes: the first on bit 6, the second on bit 5...
func smartPortEncodeHighBits(values []uint8) uint8 {
	bits := uint8(0x80)
	for i, value := range values {
		if value&0x80 != 0 {
			bits |= 0x40 >> i
		}
	}
	return bits
}

/*
smartPortDecodePacket decodes a packet, starting with the $C3 of the sync.
It returns its destination, its type and its data, and whether the checksum
and the end mark are right.
*/
func smartPortDecodePacket(packet []uint8) (uint8, uint8, []uint8, bool) {
	header := packet[1:8]
	odd := int(header[5] & 0x7f)
	groups := int(header[6] & 0x7f)

	checksum := uint8(0)
	for _, value := range header {
		checksum ^= value
	}

	var data []uint8
	position := 8
	decode := func(count int) {
		bits := packet[position]
		position++
		for i := range count {
			value := packet[position] & 0x7f
			if bits&(0x40>>i) != 0 {
				value |= 0x80
			}
			data = append(data, value)
			checksum ^= value
			position++
		}
	}
	if odd > 0 {
		decode(odd)
	}
	for range groups {
		decode(7)
	}

	received := packet[position] & (packet[position+1]<<1 | 1)
	ok := received == checksum && packet[position+2] == smartPortBusEndMark
	return header[0] & 0x7f, header[2], data, ok
}
