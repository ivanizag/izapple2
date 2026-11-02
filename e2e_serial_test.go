package izapple2

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// basicWithSerial starts a machine on BASIC with the serial card overrides
func basicWithSerial(t *testing.T, model string, overrides map[string]string) (*Apple2, *strobedKeys) {
	t.Helper()
	iic := strings.HasPrefix(model, "2c")
	if iic {
		overrides["s6"] = "iicdisk" // The disk port without diskettes
	} else {
		overrides["s6"] = "empty"
	}
	a, err := CreateAppleFromModel(model, overrides, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := &strobedKeys{}
	a.SetKeyboardProvider(keys)
	a.Init()
	a.RunCycles(2_000_000)
	if iic {
		// Without a disk the firmware waits for one, Ctrl-Reset goes to BASIC
		a.reset()
		a.RunCycles(1_000_000)
	}
	return a, keys
}

// freeTCPPort returns a port nobody listens on
func freeTCPPort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
}

// connectSerial connects to the TCP port of a serial device, once the
// device accepts it
func connectSerial(t *testing.T, a *Apple2, port string) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	// The device is ready once the connection is accepted and seen
	time.Sleep(100 * time.Millisecond)
	a.RunCycles(100_000)
	return conn
}

// expectPrinted looks for a text sent to a file. With 8 data bits the
// firmware sends the characters with the high bit set, as they are on the
// Apple II.
func expectPrinted(t *testing.T, filename string, text string) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for i := range data {
		data[i] &= 0x7f
	}
	if !strings.Contains(string(data), text) {
		t.Errorf("expected %q on the serial port, got %q", text, data)
	}
}

func TestSuperSerialPrintsToFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "serial.out")
	a, keys := basicWithSerial(t, "2plus", map[string]string{
		"s2": "ssc,device=file,file=\"" + out + "\"",
	})
	typeLines(t, a, keys, "PR#2", `PRINT "HELLO FROM THE SSC"`)
	a.RunCycles(1_000_000)
	expectPrinted(t, out, "HELLO FROM THE SSC")
}

func TestSuperSerialReceivesFromTCP(t *testing.T) {
	port := freeTCPPort(t)
	a, keys := basicWithSerial(t, "2plus", map[string]string{
		"s2": "ssc,device=tcp,port=" + port,
	})
	conn := connectSerial(t, a, port)
	typeLines(t, a, keys, "IN#2")
	a.RunCycles(1_000_000)

	if _, err := conn.Write([]byte("PRINT 6*7\r")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	cyclesToText(t, a, "42", a.GetCycles()+20_000_000)
}

func TestSuperSerialSwitches(t *testing.T) {
	a, err := CreateAppleFromModel("2plus", map[string]string{
		"s2": "ssc,baud=1200,mode=printer,stopbits=2,crdelay,width=132,linefeed",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	c := a.cards[2].(*CardSuperSerial)
	// Switches on read as 0, the Z bits as 1
	if c.sw1 != 0x8e {
		t.Errorf("expected SW1 $8e for 1200 baud and printer mode, got $%02x", c.sw1)
	}
	if c.sw2 != 0xdc {
		t.Errorf("expected SW2 $dc for 2 stop bits, CR delay, 132 columns and LF, got $%02x", c.sw2)
	}
}

func Test2cSerialPrintsToFile(t *testing.T) {
	for _, model := range iicModels {
		t.Run(model, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "printer.out")
			a, keys := basicWithSerial(t, model, map[string]string{
				"s1": "iicserial,device=file,file=\"" + out + "\"",
			})
			typeLines(t, a, keys, "PR#1", `PRINT "HELLO FROM THE IIC"`)
			a.RunCycles(1_000_000)
			expectPrinted(t, out, "HELLO FROM THE IIC")
		})
	}
}

func Test2cSerialReceivesFromTCP(t *testing.T) {
	for _, model := range iicModels {
		t.Run(model, func(t *testing.T) {
			port := freeTCPPort(t)
			a, keys := basicWithSerial(t, model, map[string]string{
				"s2": "iicserial,device=tcp,port=" + port,
			})
			conn := connectSerial(t, a, port)
			typeLines(t, a, keys, "IN#2")
			a.RunCycles(1_000_000)

			if _, err := conn.Write([]byte("PRINT 6*7\r")); err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
			cyclesToText(t, a, "42", a.GetCycles()+20_000_000)
		})
	}
}

func Test2cSerialRefusesEmptyPorts(t *testing.T) {
	_, err := CreateAppleFromModel("2c", map[string]string{"s2": "empty"}, nil)
	if err == nil {
		t.Error("expected the IIc to refuse the port 2 removed")
	}
}

func Test2cSerialOnlyOnTheIIc(t *testing.T) {
	_, err := CreateAppleFromModel("2enh", map[string]string{"s2": "iicserial"}, nil)
	if err == nil {
		t.Error("expected the IIe to refuse the IIc serial port")
	}
}

/*
keyboardBufferingProgram turns on the keyboard buffering of the firmware as
in Appendix E of the Apple IIc Technical Reference Manual. It waits for $0390
to be set and reads three keys with RDKEY to $0391-$0393. On the memory
expansion IIc, the ROMs 3 and 4, the screen holes $nnnF are $nnnC.
*/
func keyboardBufferingProgram(a *Apple2) []uint8 {
	hole := uint8(0xff)
	if version := a.mmu.Peek(0xfbbf); version == 3 || version == 4 {
		hole = 0xfc
	}
	return []uint8{
		0x78,       // SEI
		0xa9, 0x80, // LDA #$80
		0x8d, 0xfa, 0x05, // STA $05FA, buffer without the user handler
		0x8d, hole, 0x05, // STA $05FF, buffer pointers
		0x8d, hole, 0x06, // STA $06FF
		0xad, 0xaa, 0xc0, // LDA $C0AA, ACIA 2 command
		0x09, 0x0f, // ORA #$0F
		0x8d, 0xaa, 0xc0, // STA $C0AA
		0x58,             // CLI
		0xad, 0x90, 0x03, // LDA $0390
		0xf0, 0xfb, // BEQ to LDA $0390
		0x20, 0x0c, 0xfd, // JSR $FD0C, RDKEY
		0x8d, 0x91, 0x03, // STA $0391
		0x20, 0x0c, 0xfd, // JSR $FD0C
		0x8d, 0x92, 0x03, // STA $0392
		0x20, 0x0c, 0xfd, // JSR $FD0C
		0x8d, 0x93, 0x03, // STA $0393
		0x80, 0xfe, // BRA *
	}
}

func Test2cKeyboardBuffering(t *testing.T) {
	for _, model := range iicModels {
		t.Run(model, func(t *testing.T) {
			a, keys := basicWithSerial(t, model, map[string]string{})
			a.mmu.Poke(0x0390, 0)
			a.mmu.pokeRange(0x0300, keyboardBufferingProgram(a))
			a.cpu.SetPC(0x0300)
			a.RunCycles(100_000)

			// The program does not read the keyboard, the firmware takes the keys
			keys.pending = []uint8("ABC")
			a.RunCycles(1_000_000)
			if len(keys.pending) != 0 {
				t.Fatalf("expected the keys buffered, %v are waiting", len(keys.pending))
			}

			a.mmu.Poke(0x0390, 1)
			a.RunCycles(1_000_000)
			got := string([]uint8{a.mmu.Peek(0x0391) & 0x7f, a.mmu.Peek(0x0392) & 0x7f, a.mmu.Peek(0x0393) & 0x7f})
			if got != "ABC" {
				t.Errorf("expected RDKEY to read ABC from the buffer, got %q", got)
			}
		})
	}
}

func Test2cNetModemDialsFromTerminalMode(t *testing.T) {
	for _, model := range iicModels {
		t.Run(model, func(t *testing.T) {
			if model == "2c_rom3" {
				// ROM 4 fixed "a problem when using the serial modem port in
				// terminal mode": on ROM 3 the keys typed come out repeated
				t.Skip("the terminal mode of the modem port fails on ROM 3")
			}
			address, conns := testServer(t)
			a, keys := basicWithSerial(t, model, map[string]string{
				"s2": "iicserial,device=netmodem",
			})
			// The terminal mode, then 7 data bits and 9600 baud
			typeLines(t, a, keys, "IN#2")
			keys.pending = append(keys.pending, []uint8("\x01T\x011D\x0114B")...)
			a.RunCycles(2_000_000)
			if control := a.cards[2].(*CardSerialIIc).port.acia.Read(3); control != 0x3e {
				t.Errorf("expected the ACIA at 9600 baud with 7 data bits, control $3e, got $%02x", control)
			}

			// The machine runs while the modem dials
			typeLines(t, a, keys, "ATDT"+address)
			var conn net.Conn
			for start := a.GetCycles(); conn == nil; {
				select {
				case conn = <-conns:
				default:
					if a.GetCycles()-start > 20_000_000 {
						t.Fatalf("the modem did not dial:\n%v", a.ScreenText())
					}
					a.RunCycles(100_000)
				}
			}
			defer conn.Close()
			conn.Write([]byte("WELCOME TO THE BBS\r\n"))
			time.Sleep(100 * time.Millisecond)
			cyclesToText(t, a, "WELCOME TO THE BBS", a.GetCycles()+20_000_000)
			if !strings.Contains(a.ScreenText(), "CONNECT") {
				t.Errorf("expected CONNECT on the screen:\n%v", a.ScreenText())
			}

			typeLines(t, a, keys, "HELLO")
			a.RunCycles(1_000_000)
			serverExpect(t, conn, []uint8("HELLO\r"))
		})
	}
}

func TestSuperSerialNetModemDialsFromTerminalMode(t *testing.T) {
	address, conns := testServer(t)
	a, keys := basicWithSerial(t, "2plus", map[string]string{
		"s2": "ssc,device=netmodem,databits=7",
	})
	typeLines(t, a, keys, "IN#2")
	keys.pending = append(keys.pending, []uint8("\x01T\r")...)
	a.RunCycles(2_000_000)

	typeLines(t, a, keys, "ATDT"+address)
	var conn net.Conn
	for start := a.GetCycles(); conn == nil; {
		select {
		case conn = <-conns:
		default:
			if a.GetCycles()-start > 20_000_000 {
				t.Fatalf("the modem did not dial:\n%v", a.ScreenText())
			}
			a.RunCycles(100_000)
		}
	}
	defer conn.Close()
	conn.Write([]byte("WELCOME TO THE BBS\r\n"))
	time.Sleep(100 * time.Millisecond)
	cyclesToText(t, a, "WELCOME TO THE BBS", a.GetCycles()+20_000_000)

	typeLines(t, a, keys, "HELLO")
	a.RunCycles(1_000_000)
	serverExpect(t, conn, []uint8("HELLO\r"))
}
