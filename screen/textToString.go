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

	content := ""
	for l := 0; l < textLines; l++ {
		line := ""
		for c := 0; c < columns; c++ {
			char := text[l*columns+c]
			line += textMemoryByteToString(char, isAltText, supportsLowercase, false)
		}
		line = strings.TrimRight(line, " ")
		content += fmt.Sprintf("%v\n", line)
	}
	return content
}
