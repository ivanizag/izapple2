package main

import (
	"fmt"

	"github.com/ivanizag/izapple2/component"
	"github.com/ivanizag/izapple2/storage"
)

// https://embeddedmicro.weebly.com/apple-ii-p6-prom-dump.html
func getInstruction(data []uint8, Q6 bool, Q7 bool, high bool, pulse bool, sequence uint8) uint8 {

	seqBits := component.ByteToPins(sequence)
	address := component.PinsToByte([8]bool{
		seqBits[1],
		high,
		Q6,
		Q7,
		!pulse,
		seqBits[0],
		seqBits[2],
		seqBits[3],
	})
	v := data[address]
	//fmt.Printf("Seq %x. pulse %v, P6ROM[%02x]=%02x", sequence, pulse, i, v)

	fmt.Printf("For Q6(%v) Q7(%v) H(%v) P(%v) Seq(%x) => ",
		Q6, Q7, high, pulse, sequence,
	)

	return v
}

func printTable(data []uint8, q6 bool, q7 bool) {
	for i := uint8(0); i < 16; i++ {
		fmt.Printf("%x  %02x  %02x  %02x  %02x\n",
			i,
			getInstruction(data, q6, q7, false, true, i),
			getInstruction(data, q6, q7, false, false, i),
			getInstruction(data, q6, q7, true, true, i),
			getInstruction(data, q6, q7, true, false, i),
		)
	}
}

func mainXX() {
	fmt.Println("SM Analysis:")
	data, _, _ := storage.LoadResource("<internal>/DISK2P6.rom")

	fmt.Println("Q6 On, Q7 Off")
	printTable(data, true, false)
	fmt.Println("Q6 Off, Q7 Off")
	printTable(data, false, false)
	fmt.Println("Q6 On, Q7 On")
	printTable(data, true, true)
	fmt.Println("Q6 Off, Q7 On")
	printTable(data, false, true)

}

type lsq struct {
	sequence uint8
	register uint8 // 74LS323, https://components101.com/asset/sites/default/files/component_datasheet/SN74LS323-Datasheet.pdf
}

func main() {
	/*
		$C08E -> q7 false
		Every time $C08C high byte is set -> q6 false
	*/
	data, _, _ := storage.LoadResource("<internal>/DISK2P6.rom")

	wozData, _, err := storage.LoadResource("/home/casa/applerelease/disks/woz test images/WOZ 2.0/DOS 3.3 System Master.woz")
	if err != nil {
		panic("Error creando woz")
	}
	woz, err := storage.NewFileWoz(wozData)

	var lsq lsq

	nibs := woz.DumpTrackAsNib(0)
	for i := 0; i < 20; i++ {
		fmt.Printf("\n%v: ", i)
		for j := 0; j < 16; j++ {
			fmt.Printf("%02x ", nibs[i*16+j])
		}
	}
	fmt.Println("")

	/*
		bits := woz.DumpTrackAsWoz(0)
		for i := 0; i < 20; i++ {
			fmt.Printf("%v: %02x\n", i, bits[i])
		}
	*/

	position := uint32(1)
	maxPosition := uint32(50304)
	hcycles := 0
	lastReadCycle := 0
	bit := uint8(0)
	for position != 0 {
		hcycles++
		pulse := false
		if (hcycles % 8) == 0 {
			bit, position, maxPosition = woz.GetNextBitAndPosition(position, maxPosition, 0)
			pulse = bit != 0
			fmt.Printf("==Read bit %v @ %v \n", bit, position)
		}
		high := (lsq.register & 0x80) != 0
		command := getInstruction(data, false, false, high, pulse, lsq.sequence)
		//fmt.Printf(" reg %02x\n", lsq.register)
		inst := command & 0xf
		next := command >> 4
		fmt.Printf("cmd(%x) seq(%x)\n", inst, next)

		/*

			https://www.youtube.com/watch?v=r1VlrJboDMw 21:30


					   74LS323 pins: (register)
					   SR = inst[3]
					   S1 = inst[0]
					   S0 = inst[1]
					   DS0 = write protect
					   DS7 = inst[2]
					   IO[7.0] = D[7-0] slot data bus
					   OE1 = AD[0], register copied to bus depending on this
					   OE2 = (DEV pin 41 slot)
					   Q7 = high
					   Q0 = not used
					   CP = clock pulse (Q3 pin37 slot, 2Mhz clock) AND motor (using 74LS132)

					   ROM P6 pins:
					   A0, A5, A6, A7 <= sequence from 74LS174
					   D4, D5, D6, D7 => sequence to 74LS174
					   A4 <= pulse transition
					   A2 <= Q6 from 9334
					   A3 <= Q7 from 9334
					   A1 =< MSB of register (pin Q7)
					   D0-S3 => S1, S0, SR and DS7 of register
					   E2: motor

					   74LS174 pins (hex flip flop) (sequence 0,1,2,5)
					   CP = clock pulse (Q3 pin37 slot, 2Mhz clock) AND motor (using 74LS132)
					   MR = motor
					   seq:
					   	D0 <- ROM D7
					   	Q0 -> ROM A7
					   	D1 <- ROM D6
					   	Q1 -> ROM A6
					   	D2 <- ROM D4
					   	Q2 -> ROM A5
					   	D5 <- ROM D5
					   	Q5 -> ROM A0
					   pulse transition:
					   	D3 <- from Q4
					   	Q3 -> and not Q4 -> to ROM A4 (detects change in pulse)
					   	D4 <- pulse (from disk)
					   	Q4 -> to D3

					   9334 (or 74LS259) 8 bit latch
					   Q0-Q3: phases
					   Q4: motor control to 556
					   Q5: drive select
					   Q6: to ROM P6
					   Q7: to ROM P7
					   R: to slot reset
					   A0: slot AD[1]
					   A1: slot AD[2]
					   A3: slot AD[3]
					   D: slot AD[0]
					   E: (DEV 41 pin slot)


					   Write on Q7 from the 9334 and Q0 (seq[3]) transition on 74ls174

						556, for the motor signal, to continue a bit longer?

		*/

		//fmt.Printf("cmd %02x, reg %02x, inst %x, seq %x, next %x, pulse %v  ", command, lsq.register, inst, lsq.sequence, next, pulse)
		switch inst {
		case 0: // CLR, 74LS323 pin 9, Sync Reset
			lsq.register = 0
			//fmt.Printf("CLR")
		case 8: // NOP
			// Nothing
			//fmt.Printf("NOP")
		case 9: // SL0 -> S1=1, S0=0
			lsq.register = lsq.register << 1
			//fmt.Printf("SL0")
		case 0xa: // SR
			lsq.register = lsq.register >> 1 // + write protect MSB on pin 11 DS0
			//fmt.Printf("SR")
		case 0xb: // LD -> parallel load S1=1, S0=1
			panic("not in read mode")
			// lsq.register = bus
		case 0xd: // SL1
			lsq.register = 1 + (lsq.register << 1)
			//fmt.Printf("SL1")
		default:
			panic("missing instruction")
		}
		lsq.sequence = next
		//fmt.Println("")
		if lsq.register > 0x7f && (hcycles-lastReadCycle) > 60 {
			fmt.Printf("Byte %02x at %v\n", lsq.register, position)
			lastReadCycle = hcycles
		}
	}

}
