package fujinet

import (
	"encoding/json"
	"strconv"
	"strings"
)

/*
See:
	https://github.com/FujiNetWIFI/fujinet-platformio/tree/master/lib/fnjson
	https://github.com/FujiNetWIFI/fujinet-platformio/wiki/JSON-Query-Format
*/

type FnJson struct {
	data   any
	Result []uint8
}

func NewFnJson() *FnJson {
	var js FnJson
	return &js
}

func (js *FnJson) Parse(data []uint8) ErrorCode {
	// See FNJSON::parse()
	err := json.Unmarshal(data, &js.data)
	if err != nil {
		return NetworkErrorJsonParseError
	}
	return NoError
}

func (js *FnJson) Query(query []uint8) {
	// See FNJSON::setReadQuery
	// See https://github.com/kbranigan/cJSON/blob/master/cJSON_Utils.c
	if query[len(query)-1] == 0 {
		query = query[0 : len(query)-1]
	}
	queryString := string(query)
	queryString = strings.TrimSuffix(queryString, "/0")
	queryString = strings.TrimPrefix(queryString, "/")
	queryString = strings.TrimSuffix(queryString, "/")
	path := strings.Split(queryString, "/")

	js.Result = getJsonValue(nil)
	current := js.data
	for i := 0; i < len(path); i++ {
		switch v := current.(type) {
		case map[string]any:
			var found bool
			current, found = v[path[i]]
			if !found {
				// Not found
				return
			}
		case []any:
			index, err := strconv.Atoi(path[i])
			if err != nil {
				// Path for arrays should be an int
				return
			}
			if index < 0 || index >= len(v) {
				// Index out of bounds
				return
			}
			current = v[index]
		default:
			// It's a leaf. We can't go down
			return
		}
	}

	js.Result = getJsonValue(current)
}

func getJsonValue(data any) []uint8 {
	// See FNJson::getValue
	if data == nil {
		return []uint8("NULL")
	}

	switch v := data.(type) {
	case bool:
		if v {
			return []uint8("TRUE")
		} else {
			return []uint8("FALSE")
		}
	case float64:
		// if math.Floor(v) == v { // As done in FNJSON__getValue()
		//  It's an integer
		return []uint8(strconv.Itoa(int(v)))
		// } else {
		//	 return []uint8(fmt.Sprintf("%.10f", v))
		// }
	case string:
		return []uint8(v)
	case []any:
		s := make([]uint8, 0)
		for i := 0; i < len(v); i++ {
			s = append(s, getJsonValue(v[i])...)
		}
		return s
	case map[string]any:
		s := make([]uint8, 0)
		for k, e := range v {
			s = append(s, []uint8(k)...)
			s = append(s, getJsonValue(e)...)
		}
		return s
	default:
		// Should not be possible for an object unmarshalled from a JSON
		return []uint8("UNKNOWN")
	}
}
