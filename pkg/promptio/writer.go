package promptio

import prompt "github.com/c-bata/go-prompt"

// StableAreaWriter wraps a go-prompt ConsoleWriter so that prepareArea
// (used to reserve space for the completion/description panel) does not
// emit VT100 IND/RI (ESC D / ESC M).
//
// Those sequences are how stock go-prompt implements ScrollDown/ScrollUp.
// On Windows they are the reason the help panel drifts upward:
//
//   - A native Windows build writes through mattn/go-colorable, which
//     silently drops ESC D and ESC M, so no space is reserved. Drawing the
//     panel at the bottom of conhost then scrolls the viewport.
//   - The same sequences reach W10 Cmd/PowerShell when medit runs on
//     Android/Linux and the host is adb, OpenSSH, etc. IND at the last row
//     scrolls content up; RI does not reverse that scroll. Each keystroke
//     therefore lifts the panel by one or more lines until it covers the
//     input line (https://github.com/sterrasec/apk-medit/issues/37).
//
// Instead, the first ScrollDown saves the cursor, each ScrollDown writes a
// newline to actually allocate a row, and the matching last ScrollUp
// restores the cursor. ESC [s / ESC [u are handled by go-colorable and by
// Windows VT processing, so the prompt stays put on Windows and on Unix.
//
// DropBareNewlines, when set, swallows the extra lone '\n' go-prompt emits
// from lineWrap on non-Windows GOOS. Windows consoles already wrap at the
// last column; the extra newline is another source of upward drift when a
// Linux/Android binary is viewed from Cmd or PowerShell.
type StableAreaWriter struct {
	prompt.ConsoleWriter
	reserved         int
	allocating       bool
	DropBareNewlines bool
}

// NewStableAreaWriter returns a ConsoleWriter that reserves completion-panel
// space without IND/RI scrolling.
func NewStableAreaWriter(inner prompt.ConsoleWriter) *StableAreaWriter {
	if inner == nil {
		inner = prompt.NewStdoutWriter()
	}
	return &StableAreaWriter{ConsoleWriter: inner}
}

// WriteRaw implements prompt.ConsoleWriter.
func (w *StableAreaWriter) WriteRaw(data []byte) {
	if w.DropBareNewlines && !w.allocating && len(data) == 1 && data[0] == '\n' {
		return
	}
	w.ConsoleWriter.WriteRaw(data)
}

// ScrollDown allocates one row below the cursor for the suggestion panel.
func (w *StableAreaWriter) ScrollDown() {
	if w.reserved == 0 {
		w.SaveCursor()
	}
	w.allocating = true
	w.WriteRaw([]byte{'\n'})
	w.allocating = false
	w.reserved++
}

// ScrollUp pairs with ScrollDown after space has been reserved.
func (w *StableAreaWriter) ScrollUp() {
	if w.reserved > 0 {
		w.reserved--
		if w.reserved == 0 {
			w.UnSaveCursor()
		}
		return
	}
	w.ConsoleWriter.ScrollUp()
}
