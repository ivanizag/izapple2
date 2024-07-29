package izapple2

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ivanizag/izapple2/fujinet"
)

/*

The network device as implemented by Fujinet:

See:
	https://github.com/FujiNetWIFI/fujinet-platformio/blob/master/lib/device/iwm/network.cpp

*/

// SmartPortFujinetNetwork represents a Fujinet device
type SmartPortFujinetNetwork struct {
	host  *CardSmartPort // For DMA
	trace bool

	protocol        fujinet.Protocol
	jsonChannelMode bool
	statusByte      uint8
	errorCode       fujinet.ErrorCode

	jsonData *fujinet.FnJson
	data     []uint8
	// connected    uint8
}

// NewSmartPortFujinetNetwork creates a new fujinet device
func NewSmartPortFujinetNetwork(host *CardSmartPort) *SmartPortFujinetNetwork {
	var d SmartPortFujinetNetwork
	d.host = host
	d.errorCode = fujinet.NoError

	return &d
}

func (d *SmartPortFujinetNetwork) exec(call *smartPortCall) uint8 {
	var result uint8

	switch call.command {

	case smartPortCommandOpen:
		result = smartPortNoError

	case smartPortCommandClose:
		result = smartPortNoError

	case smartPortCommandStatus:
		address := call.param16(2)
		result = d.status(call.statusCode(), address)

	case smartPortCommandControl:
		data := call.paramData(2)
		controlCode := call.param8(4)
		result = d.control(data, controlCode)

	case smartPortCommandRead:
		address := call.param16(2)
		len := call.param16(4)
		pos := call.param24(6)
		result = d.read(pos, len, address)

	default:
		// Prodos device command not supported
		result = smartPortErrorIO
	}

	if d.trace {
		fmt.Printf("[SmartPortFujinetNetwork] Command %v, return %s \n",
			call, smartPortErrorMessage(result))
	}

	return result
}

func (d *SmartPortFujinetNetwork) read(pos uint32, length uint16, dest uint16) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortFujinetNetwork] Read %v bytes from pos %v into $%x.\n",
			length, pos, dest)
	}

	// Byte by byte transfer to memory using the full Poke code path
	for i := uint16(0); i < uint16(len(d.data)) && i < length; i++ {
		d.host.a.mmu.Poke(dest+i, d.data[i])
	}

	return smartPortNoError
}

func (d *SmartPortFujinetNetwork) control(data []uint8, code uint8) uint8 {
	switch code {
	case 'O':
		// Open URL
		method := data[0]
		translation := data[1]
		url := data[2:]
		d.controlOpen(method, translation, string(url))

	case 'P':
		if d.jsonChannelMode {
			d.controlJsonParse()
		}

	case 'Q':
		if d.jsonChannelMode {
			d.controlJsonQuery(data)
		}

	case 0xfc:
		mode := data[0]
		d.controlChannelMode(mode)
	}

	return smartPortNoError
}

func (d *SmartPortFujinetNetwork) controlJsonParse() {
	// See FNJSON::parse()
	if d.trace {
		fmt.Printf("[SmartPortFujinetNetwork] control-parse()\n")
	}

	data, errorCode := d.protocol.ReadAll()
	if errorCode != fujinet.NoError {
		d.errorCode = errorCode
		return
	}

	d.jsonData = fujinet.NewFnJson()
	d.errorCode = d.jsonData.Parse(data)
}

func (d *SmartPortFujinetNetwork) controlJsonQuery(query []uint8) {
	if d.trace {
		fmt.Printf("[SmartPortFujinetNetwork] control-query('%s')\n", query)
	}

	if d.jsonData != nil {
		d.jsonData.Query(query)
		d.data = d.jsonData.Result
	}
}

func (d *SmartPortFujinetNetwork) controlChannelMode(mode uint8) {
	// See iwmNetwork::channel_mode()
	if d.trace {
		fmt.Printf("control-channel-mode(%v)\n", mode)
	}

	if mode == 0 {
		d.jsonChannelMode = false
	} else if mode == 1 {
		d.jsonChannelMode = true
	}
	// The rest of the cases do not change the mode
}

func (d *SmartPortFujinetNetwork) controlOpen(method uint8, translation uint8, rawUrl string) {
	// See iwmNetwork::open()
	if d.trace {
		fmt.Printf("[SmartPortFujinetNetwork] control-open(%v, %v, '%s'\n", method, translation, rawUrl)
	}

	if d.protocol != nil {
		d.protocol.Close()
		d.protocol = nil
	}
	d.statusByte = 0

	// Remove "N:" prefix
	rawUrl = strings.TrimPrefix(rawUrl, "N:")

	urlParsed, err := url.Parse(rawUrl)
	if err != nil {
		d.errorCode = fujinet.NetworkErrorInvalidDeviceSpec
		d.statusByte = 4 // client_error
	}

	d.protocol, d.errorCode = fujinet.InstantiateProtocol(urlParsed, method)
	if d.protocol == nil {
		d.statusByte = 4 // client_error
		return
	}

	d.protocol.Open(urlParsed)
	d.jsonChannelMode = false
}

func (d *SmartPortFujinetNetwork) status(code uint8, dest uint16) uint8 {

	switch code {
	case smartPortStatusCodeDevice:
		// See iwmNetwork::encode_status_reply_packet()
		d.host.a.mmu.pokeRange(dest, []uint8{
			smartPortStatusCodeTypeRead & smartPortStatusCodeTypeOnline,
			0, 0, 0, // Block size is 0
		})

	case smartPortStatusCodeDeviceInfo:
		// See iwmNetwork::encode_status_reply_packet()
		d.host.a.mmu.pokeRange(dest, []uint8{
			smartPortStatusCodeTypeRead & smartPortStatusCodeTypeOnline,
			0, 0, 0, // Block size is 0
			7, 'N', 'E', 'T', 'W', 'O', 'R', 'K', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
			0x02,       // Type hard disk
			0x00,       // Subtype network (comment in network.cpp has 0x0a)
			0x00, 0x01, // Firmware version
		})

	case 'R':
		// Net read, do nothing

	case 'S':
		// Get connection status
		len := len(d.data)
		if d.jsonChannelMode {
			// See FNJSON
			errorCode := 0
			if len == 0 {
				errorCode = int(fujinet.NetworkErrorEndOfFile)
			}
			d.host.a.mmu.pokeRange(dest, []uint8{
				uint8(len & 0xff),
				uint8((len >> 8) & 0xff),
				1, /*True*/
				uint8(errorCode),
			})
		} else {
			// TODO
			d.host.a.mmu.pokeRange(dest, []uint8{
				uint8(len & 0xff),
				uint8((len >> 8) & 0xff),
				1, // ?? d.connected,
				uint8(d.errorCode),
			})
		}
	}

	return smartPortNoError // The return code is always success
}
