package apple2

import "testing"

func TestCharAddress(t *testing.T) {
	var tp textPages

	mappings := [][]uint8{
		// column, line, page, address
		{0, 0, 0, 0},
		{0, 1, 0, 0x80},
		{0, 2, 1, 0x00},
		{0, 23, 3, 0xD0},
	}

	for _, v := range mappings {
		page, address := tp.charAddress(v[0], v[1])
		if page != v[2] || address != v[3] {
			t.Errorf("Error on charAddress for (%v, %v) (%v:%02x) <> (%v:%02x)",
				v[0], v[1], v[2], v[3], page, address)
		}
	}
}
