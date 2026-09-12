package promptio

import (
	"fmt"
	"testing"
)

// windowsLikeConsole models W10 Cmd/PowerShell (conhost) well enough to
// show why go-prompt's IND/RI prepareArea drifts and why save+newline+restore
// does not.
//
// Rules that match the reported "rises one or more lines per keystroke":
//   - IND (ESC D) on the last row scrolls the viewport up; the cursor stays.
//   - RI  (ESC M) only moves the cursor up; it does not reverse a scroll.
//   - ESC [s / ESC [u save and restore buffer coordinates.
//   - A newline at the last row allocates a blank row (scrolls if needed)
//     then restore puts the cursor back on the prompt row.
type windowsLikeConsole struct {
	height int
	// lines[0] is the top of the viewport. Each entry is a marker so we
	// can see whether the prompt row has been pushed off-screen.
	lines    []string
	curY     int
	savedY   int
	hasSaved bool
}

func newWindowsLikeConsole(height int, promptRow string) *windowsLikeConsole {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = fmt.Sprintf("out%d", i)
	}
	// Prompt sits on the last visible row — the usual state after medit
	// prints process-list output and then opens "> ".
	lines[height-1] = promptRow
	return &windowsLikeConsole{
		height: height,
		lines:  lines,
		curY:   height - 1,
	}
}

func (c *windowsLikeConsole) scrollUp() {
	copy(c.lines, c.lines[1:])
	c.lines[c.height-1] = ""
}

func (c *windowsLikeConsole) ind() {
	if c.curY >= c.height-1 {
		c.scrollUp()
		return
	}
	c.curY++
}

func (c *windowsLikeConsole) ri() {
	if c.curY > 0 {
		c.curY--
	}
}

func (c *windowsLikeConsole) save() {
	c.savedY = c.curY
	c.hasSaved = true
}

func (c *windowsLikeConsole) restore() {
	if c.hasSaved {
		c.curY = c.savedY
	}
}

func (c *windowsLikeConsole) newline() {
	if c.curY >= c.height-1 {
		c.scrollUp()
		if c.hasSaved {
			c.savedY--
			if c.savedY < 0 {
				c.savedY = 0
			}
		}
		return
	}
	c.curY++
}

func (c *windowsLikeConsole) promptRow() int {
	for i, line := range c.lines {
		if line == "> " {
			return i
		}
	}
	return -1
}

func prepareAreaIND(c *windowsLikeConsole, n int) {
	for i := 0; i < n; i++ {
		c.ind()
	}
	for i := 0; i < n; i++ {
		c.ri()
	}
}

func prepareAreaStable(c *windowsLikeConsole, n int) {
	c.save()
	for i := 0; i < n; i++ {
		c.newline()
	}
	c.restore()
}

func TestWindowsINDPrepareAreaDriftsUpward(t *testing.T) {
	c := newWindowsLikeConsole(10, "> ")
	start := c.promptRow()
	if start != 9 {
		t.Fatalf("prompt should start on last row, got %d", start)
	}

	prepareAreaIND(c, 4)
	row := c.promptRow()
	if row < 0 {
		t.Fatal("prompt scrolled out of the viewport")
	}
	if row >= start {
		t.Fatalf("IND at the last row should lift the prompt; still at row %d", row)
	}
}

func TestNoopPrepareAreaPlusLineWrapDriftsEveryKeystroke(t *testing.T) {
	// go-colorable drops IND/RI, so prepareArea is a no-op. go-prompt then
	// draws N panel rows and (on a Linux/Android binary) emits an extra
	// lineWrap newline per row. CursorUp only accounts for N rows, so the
	// prompt rises on every keystroke — the reporter's "flying blind" bug.
	c := newWindowsLikeConsole(12, "> ")
	start := c.promptRow()
	const panel = 3
	for keystroke := 1; keystroke <= 2; keystroke++ {
		for i := 0; i < panel; i++ {
			c.newline() // row write that may scroll
			c.newline() // extra lineWrap \n
		}
		for i := 0; i < panel; i++ {
			c.ri()
		}
		row := c.promptRow()
		if row < 0 {
			t.Fatalf("keystroke %d: prompt left the viewport", keystroke)
		}
		if row >= start {
			t.Fatalf("keystroke %d: expected prompt to rise from %d, still at %d", keystroke, start, row)
		}
		start = row
	}
}

func TestWindowsStablePrepareAreaKeepsPromptVisible(t *testing.T) {
	c := newWindowsLikeConsole(10, "> ")
	start := c.promptRow()

	const panel = 4
	for keystroke := 1; keystroke <= 8; keystroke++ {
		prepareAreaStable(c, panel)
		row := c.promptRow()
		if row < 0 {
			t.Fatalf("keystroke %d: prompt missing after stable prepareArea", keystroke)
		}
		// Reserving space may move the prompt up once (to make room for
		// the panel at the bottom), but it must not keep walking upward.
		if keystroke == 1 {
			if row > start {
				t.Fatalf("prompt moved down to %d", row)
			}
			start = row
			continue
		}
		if row != start {
			t.Fatalf("keystroke %d: prompt drifted from row %d to %d", keystroke, start, row)
		}
		if c.curY != row {
			t.Fatalf("keystroke %d: cursor at %d, prompt at %d", keystroke, c.curY, row)
		}
	}
}
