package screen

import (
	"fmt"
	"strings"
)

// RenderTextModeString returns the text mode contents ignoring reverse and flash
func RenderTextModeString(vs VideoSource, is80Columns bool, isSecondPage bool, isAltText bool, supportsLowercase bool, hasAltOrder bool) string {

	var text []uint8
	if is80Columns {
		text = getText80FromMemory(vs, isSecondPage, hasAltOrder)
	} else {
		text = getTextFromMemory(vs, isSecondPage, false)
	}
	columns := len(text) / textLines

	var content strings.Builder
	for l := range textLines {
		line := ""
		for c := range columns {
			char := text[l*columns+c]
			line += textMemoryByteToString(char, isAltText, supportsLowercase, false)
		}
		line = strings.TrimRight(line, " ")
		content.WriteString(fmt.Sprintf("%v\n", line))
	}
	return content.String()
}
