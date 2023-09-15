package fujinet

import (
	"io"
	"net/http"
	"net/url"
)

type protocolHttp struct {
	method uint8
	url    *url.URL
}

func newProtocolHttp(method uint8) *protocolHttp {
	var p protocolHttp
	p.method = method
	return &p
}

func (p *protocolHttp) Open(urlParsed *url.URL) {
	p.url = urlParsed
}

func (p *protocolHttp) Close() {
	// nothing to do
}

func (p *protocolHttp) ReadAll() ([]uint8, ErrorCode) {
	if p.method == 12 /*GET*/ {
		resp, err := http.Get(p.url.String())
		if err != nil {
			return nil, NetworkErrorGeneral
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, NetworkErrorGeneral
		}
		return data, NoError
	}

	return nil, NetworkErrorNotImplemented
}

func (p *protocolHttp) Write(data []uint8) error {
	return nil
}
