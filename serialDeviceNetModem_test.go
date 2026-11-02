package izapple2

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"
)

// modemType sends a text to the modem as the Apple would
func modemType(m *serialDeviceNetModem, text string) {
	for _, value := range []uint8(text) {
		m.Send(value)
	}
}

// modemExpect reads from the modem until the text arrives
func modemExpect(t *testing.T, m *serialDeviceNetModem, text string) string {
	t.Helper()
	var received []uint8
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if value, ok := m.Receive(); ok {
			received = append(received, value)
			if strings.Contains(string(received), text) {
				return string(received)
			}
		} else {
			time.Sleep(time.Millisecond)
		}
	}
	t.Fatalf("expected %q from the modem, got %q", text, received)
	return ""
}

// testServer accepts one connection and gives it to the test
func testServer(t *testing.T) (string, chan net.Conn) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	conns := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conns <- conn
		}
	}()
	return listener.Addr().String(), conns
}

// serverExpect reads from the connection until the data arrives
func serverExpect(t *testing.T, conn net.Conn, data []uint8) {
	t.Helper()
	var received []uint8
	buffer := make([]uint8, 256)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for !bytes.Contains(received, data) {
		n, err := conn.Read(buffer)
		received = append(received, buffer[:n]...)
		if err != nil {
			t.Fatalf("expected %q on the server, got %q", data, received)
		}
	}
}

func TestNetModemCommands(t *testing.T) {
	m := newSerialDeviceNetModem()
	modemType(m, "AT\r")
	if got := modemExpect(t, m, "OK\r\n"); got != "AT\r\r\nOK\r\n" {
		t.Errorf("expected the command echoed and OK, got %q", got)
	}

	modemType(m, "ATE0\r")
	modemExpect(t, m, "OK\r\n")
	modemType(m, "ATS12=5\r")
	modemExpect(t, m, "OK\r\n")
	modemType(m, "ATS12?\r")
	modemExpect(t, m, "005")

	modemType(m, "ATV0\r")
	modemExpect(t, m, "0\r")
	modemType(m, "ATK\r")
	modemExpect(t, m, "4\r") // ERROR

	modemType(m, "ATZ\r")
	if got := modemExpect(t, m, "OK\r\n"); got != "0\r" && !strings.HasSuffix(got, "\r\nOK\r\n") {
		t.Errorf("expected OK in words after ATZ, got %q", got)
	}
	modemType(m, "A/")
	if got := modemExpect(t, m, "OK\r\n"); !strings.Contains(got, "A/") {
		t.Errorf("expected A/ to repeat ATZ with echo, got %q", got)
	}
}

func TestNetModemDialsAndHangsUp(t *testing.T) {
	address, conns := testServer(t)
	m := newSerialDeviceNetModem()
	modemType(m, "ATE0S12=5\r")
	modemExpect(t, m, "OK")

	modemType(m, "ATD"+address+"\r")
	modemExpect(t, m, "CONNECT\r\n")
	conn := <-conns
	defer conn.Close()

	conn.Write([]uint8("WELCOME"))
	modemExpect(t, m, "WELCOME")
	modemType(m, "HELLO\r")
	serverExpect(t, conn, []uint8("HELLO\r"))

	// The escape, with the guard time before and after
	time.Sleep(200 * time.Millisecond)
	modemType(m, "+++")
	modemExpect(t, m, "OK\r\n")

	// Back online, the connection is still there
	modemType(m, "ATO\r")
	modemExpect(t, m, "CONNECT")
	modemType(m, "AGAIN")
	serverExpect(t, conn, []uint8("+++AGAIN"))

	time.Sleep(200 * time.Millisecond)
	modemType(m, "+++")
	modemExpect(t, m, "OK")
	modemType(m, "ATH\r")
	modemExpect(t, m, "OK")
	// The pluses were sent, then the connection is closed
	conn.SetReadDeadline(time.Now().Add(time.Second))
	buffer := make([]uint8, 10)
	for {
		if _, err := conn.Read(buffer); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				t.Error("expected the connection closed by ATH")
			}
			break
		}
	}
}

func TestNetModemPlusesInTheData(t *testing.T) {
	address, conns := testServer(t)
	m := newSerialDeviceNetModem()
	modemType(m, "ATE0S12=5\r")
	modemType(m, "ATD"+address+"\r")
	modemExpect(t, m, "CONNECT")
	conn := <-conns
	defer conn.Close()

	// Without the guard time before, the pluses are data
	modemType(m, "A+++B")
	time.Sleep(200 * time.Millisecond)
	modemType(m, "C")
	serverExpect(t, conn, []uint8("A+++BC"))
}

func TestNetModemRemoteHangsUp(t *testing.T) {
	address, conns := testServer(t)
	m := newSerialDeviceNetModem()
	modemType(m, "ATE0\r")
	modemType(m, "ATDT\""+address+"\"\r")
	modemExpect(t, m, "CONNECT")
	conn := <-conns
	conn.Write([]uint8("BYE"))
	conn.Close()
	modemExpect(t, m, "BYE\r\nNO CARRIER\r\n")

	// In command mode again
	modemType(m, "AT\r")
	modemExpect(t, m, "OK")
}

func TestNetModemNoAnswer(t *testing.T) {
	// A port nobody listens on
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()

	m := newSerialDeviceNetModem()
	modemType(m, "ATE0\rATD"+address+"\r")
	modemExpect(t, m, "NO CARRIER")
}

func TestNetModemTelnet(t *testing.T) {
	address, conns := testServer(t)
	m := newSerialDeviceNetModem()
	modemType(m, "ATE0\r")
	modemType(m, "ATDT"+address+"\r")
	modemExpect(t, m, "CONNECT")
	conn := <-conns
	defer conn.Close()

	// The negotiation is answered and taken out of the data
	conn.Write([]uint8{telnetIAC, telnetDO, 24, 'A', telnetIAC, telnetWILL, telnetOptionEcho,
		telnetIAC, telnetSB, 24, 1, telnetIAC, telnetSE, 'B', telnetIAC, telnetIAC, 'C'})
	if got := modemExpect(t, m, "C"); !strings.HasSuffix(got, "AB\xffC") {
		t.Errorf("expected the data without the telnet commands, got %q", got)
	}
	serverExpect(t, conn, []uint8{telnetIAC, telnetWONT, 24, telnetIAC, telnetDO, telnetOptionEcho})

	// IAC doubled, CR as CR NUL
	modemType(m, "\xff\r")
	serverExpect(t, conn, []uint8{telnetIAC, telnetIAC, '\r', 0})
}

func TestNetModemRawDial(t *testing.T) {
	address, conns := testServer(t)
	m := newSerialDeviceNetModem()
	modemType(m, "ATE0\r")
	modemType(m, "ATD"+address+"\r")
	modemExpect(t, m, "CONNECT")
	conn := <-conns
	defer conn.Close()

	// Without telnet the bytes pass as they are
	conn.Write([]uint8{telnetIAC, telnetDO, 24})
	modemExpect(t, m, "\xff\xfd\x18")
}

func TestNetModemConnectsAtTheSpeedOfThePort(t *testing.T) {
	for _, c := range []struct {
		baud     int
		commands string
		expected string
	}{
		{9600, "ATE0\r", "\r\nCONNECT 9600\r\n"},
		{300, "ATE0\r", "\r\nCONNECT\r\n"},    // As the Hayes modems at 300 baud
		{9600, "ATE0X0\r", "\r\nCONNECT\r\n"}, // The results of the Smartmodem 300
		{2400, "ATE0V0\r", "10\r"},            // The number of CONNECT 2400
	} {
		address, conns := testServer(t)
		m := newSerialDeviceNetModem()
		m.setBaudRate(c.baud)
		modemType(m, c.commands)
		modemExpect(t, m, "0") // OK, as word or number
		for _, ok := m.Receive(); ok; _, ok = m.Receive() {
		}
		modemType(m, "ATD"+address+"\r")
		conn := <-conns
		got := modemExpect(t, m, c.expected)
		if got != c.expected {
			t.Errorf("at %v baud with %q, expected %q, got %q", c.baud, c.commands, c.expected, got)
		}
		conn.Close()
	}
}
