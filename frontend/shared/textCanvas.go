package shared

import "strings"

/*
A grid of characters to compose the screens the frontends show with
screen.SnapshotMessageGenerator, like the drop targets. Everything written
outside the grid is discarded, so the callers do not have to check the bounds
of what they place.
*/

type textCanvas struct {
	columns int
	lines   int
	chars   []byte
}

func newTextCanvas(columns int, lines int) *textCanvas {
	var c textCanvas
	c.columns = columns
	c.lines = lines
	c.chars = make([]byte, columns*lines)
	for i := range c.chars {
		c.chars[i] = ' '
	}
	return &c
}

func (c *textCanvas) setChar(column int, line int, char byte) {
	if column < 0 || column >= c.columns || line < 0 || line >= c.lines {
		return
	}
	c.chars[line*c.columns+column] = char
}

// putText writes a text from a column to the right
func (c *textCanvas) putText(column int, line int, text string) {
	for i := range len(text) {
		c.setChar(column+i, line, text[i])
	}
}

// putTextCentered writes a text centered between two columns, the right one
// not included
func (c *textCanvas) putTextCentered(left int, right int, line int, text string) {
	c.putText(left+(right-left-len(text))/2, line, text)
}

// putHorizontalLine fills a line with a character between two columns, the
// right one not included
func (c *textCanvas) putHorizontalLine(left int, right int, line int, char byte) {
	for column := left; column < right; column++ {
		c.setChar(column, line, char)
	}
}

// putVerticalLine fills a column with a character between two lines, the
// bottom one not included
func (c *textCanvas) putVerticalLine(column int, top int, bottom int, char byte) {
	for line := top; line < bottom; line++ {
		c.setChar(column, line, char)
	}
}

func (c *textCanvas) String() string {
	lines := make([]string, c.lines)
	for line := range c.lines {
		lines[line] = strings.TrimRight(
			string(c.chars[line*c.columns:(line+1)*c.columns]), " ")
	}
	return strings.Join(lines, "\n")
}
