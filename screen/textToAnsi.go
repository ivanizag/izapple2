package screen

import (
	"fmt"
	"strings"
)

// RenderTextModeAnsi returns the text mode contents using ANSI escape codes for reverse and flash
func RenderTextModeAnsi(vs VideoSource, is80Columns bool, isSecondPage bool, isAltText bool, supportsLowercase bool, hasAltOrder bool) string {

	var text []uint8
	if is80Columns {
		text = getText80FromMemory(vs, isSecondPage, hasAltOrder)
	} else {
		text = getTextFromMemory(vs, isSecondPage, false)
	}
	columns := len(text) / textLines

	content := "\n"
	content += fmt.Sprintln(strings.Repeat("#", columns+4))
	for l := 0; l < textLines; l++ {
		line := ""
		for c := 0; c < columns; c++ {
			char := text[l*columns+c]
			line += textMemoryByteToString(char, isAltText, supportsLowercase, true)
		}
		content += fmt.Sprintf("# %v #\n", line)
	}

	content += fmt.Sprintln(strings.Repeat("#", columns+4))
	return content
}

/*
See Apple IIe reference manual. Table 2-5
---------Ascii----II+------IIe------AltChar--AltCEnh
$00-$1f  Control  Upp Inv  Upp Inv  Upp Inv  Upp Inv
$20-$3f  Symbols  Sym Inv  Sym Inv  Sym Inv  Upp Inv
$40-$5f  UpperCa  Upp Fla  Upp Fla  Upp Inv  Mouse
$60-$7f  LowerCa  Sym Fla  Sym Fla  Low Inv  Low Inv
$80-$9f           Upp Nor  Upp Nor  Upp Nor  Upp Nor
$a0-$bf           Sym Nor  Sym Nor  Sym Nor  Sym Nor
$c0-$df           Upp Nor  Upp Nor  Upp Nor  Upp Nor
$e0-$ff           Low Nor  Low Nor  Low Nor  Low Nor
----------------------------------------------------
*/

func textMemoryByteToString(value uint8, isAltCharSet bool, supportsLowercase bool, ansi bool) string {
	// Normal, inverse or flash
	topBits := value >> 6
	isInverse := topBits == 0
	isFlash := topBits == 1
	if isFlash && isAltCharSet {
		isFlash = false
		isInverse = true
	}

	// Move blocks
	value &= 0x7f
	if !supportsLowercase {
		// No lowercase
		value &= 0x3f
	}
	if isFlash || isInverse && !isAltCharSet {
		// No flash or inverse lowercase
		value &= 0x3f
	}
	if value < 0x20 {
		// Control is Uppercase
		value += 0x40
	}

	// Render
	if value == 0x7f {
		// DEL is full box
		value = '_'
	}

	if ansi && isFlash {
		if value == ' ' {
			// Flashing space in Apple is the full box. It can't be done with ANSI codes
			value = '_'
		}
		return fmt.Sprintf("\033[5m%v\033[0m", string(value))
	} else if ansi && isInverse {
		return fmt.Sprintf("\033[7m%v\033[0m", string(value))
	} else {
		return string(value)
	}
}
