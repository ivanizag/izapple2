package fujinet

import (
	"net/url"
	"strings"
)

type Protocol interface {
	Open(urlParsed *url.URL)
	Close()
	ReadAll() ([]uint8, ErrorCode)
	Write(data []uint8) error
}

type ErrorCode uint8

const (
	// See fujinet-platformio/lib/network-protocol/status_error_codes.h
	NoError                       = ErrorCode(0)
	NetworkErrorEndOfFile         = ErrorCode(136)
	NetworkErrorGeneral           = ErrorCode(144)
	NetworkErrorNotImplemented    = ErrorCode(146)
	NetworkErrorInvalidDeviceSpec = ErrorCode(165)

	// New
	NetworkErrorJsonParseError = ErrorCode(250)
)

func InstantiateProtocol(urlParsed *url.URL, method uint8) (Protocol, ErrorCode) {
	scheme := strings.ToUpper(urlParsed.Scheme)
	switch scheme {
	case "HTTP":
		return newProtocolHttp(method), NoError
	case "HTTPS":
		return newProtocolHttp(method), NoError
	default:
		return nil, NetworkErrorGeneral
	}
}
