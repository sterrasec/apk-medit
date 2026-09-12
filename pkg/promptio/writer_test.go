package promptio

import (
	"strings"
	"testing"

	prompt "github.com/c-bata/go-prompt"
)

// recorder logs ConsoleWriter calls that matter for prepareArea.
type recorder struct {
	prompt.ConsoleWriter
	ops []string
}

func (r *recorder) WriteRaw(data []byte) {
	r.ops = append(r.ops, "WriteRaw:"+string(data))
}
func (r *recorder) Write(data []byte)       { r.ops = append(r.ops, "Write:"+string(data)) }
func (r *recorder) WriteRawStr(data string) { r.ops = append(r.ops, "WriteRawStr:"+data) }
func (r *recorder) WriteStr(data string)    { r.ops = append(r.ops, "WriteStr:"+data) }
func (r *recorder) Flush() error            { return nil }
func (r *recorder) EraseScreen()            {}
func (r *recorder) EraseUp()                {}
func (r *recorder) EraseDown()              {}
func (r *recorder) EraseStartOfLine()       {}
func (r *recorder) EraseEndOfLine()         {}
func (r *recorder) EraseLine()              {}
func (r *recorder) ShowCursor()             {}
func (r *recorder) HideCursor()             {}
func (r *recorder) CursorGoTo(int, int)     {}
func (r *recorder) CursorUp(int)            {}
func (r *recorder) CursorDown(int)          {}
func (r *recorder) CursorForward(int)       {}
func (r *recorder) CursorBackward(int)      {}
func (r *recorder) AskForCPR()              {}
func (r *recorder) SaveCursor()             { r.ops = append(r.ops, "SaveCursor") }
func (r *recorder) UnSaveCursor()           { r.ops = append(r.ops, "UnSaveCursor") }
func (r *recorder) ScrollDown()             { r.ops = append(r.ops, "ScrollDown") }
func (r *recorder) ScrollUp()               { r.ops = append(r.ops, "ScrollUp") }
func (r *recorder) SetTitle(string)         {}
func (r *recorder) ClearTitle()             {}
func (r *recorder) SetColor(prompt.Color, prompt.Color, bool) {
}

func TestNewStableAreaWriterNilInner(t *testing.T) {
	w := NewStableAreaWriter(nil)
	if w == nil {
		t.Fatal("expected writer")
	}
}

func TestPrepareAreaUsesSaveNewlineRestore(t *testing.T) {
	inner := &recorder{}
	w := NewStableAreaWriter(inner)

	const lines = 6
	// Same loop go-prompt's Render.prepareArea uses.
	for i := 0; i < lines; i++ {
		w.ScrollDown()
	}
	for i := 0; i < lines; i++ {
		w.ScrollUp()
	}

	want := []string{"SaveCursor"}
	for i := 0; i < lines; i++ {
		want = append(want, "WriteRaw:\n")
	}
	want = append(want, "UnSaveCursor")
	if got := strings.Join(inner.ops, "|"); got != strings.Join(want, "|") {
		t.Fatalf("ops mismatch\ngot:  %q\nwant: %q", got, strings.Join(want, "|"))
	}
	if w.reserved != 0 {
		t.Fatalf("reserved=%d, want 0 after paired ScrollUp", w.reserved)
	}
	for _, op := range inner.ops {
		if op == "ScrollDown" || op == "ScrollUp" {
			t.Fatalf("inner writer received IND/RI-style scroll op %q", op)
		}
	}
}

func TestUnbalancedScrollUpFallsBack(t *testing.T) {
	inner := &recorder{}
	w := NewStableAreaWriter(inner)
	w.ScrollUp()
	if len(inner.ops) != 1 || inner.ops[0] != "ScrollUp" {
		t.Fatalf("unbalanced ScrollUp should delegate, got %v", inner.ops)
	}
}

func TestDropBareNewlinesSkipsLineWrapButKeepsPrepareArea(t *testing.T) {
	inner := &recorder{}
	w := NewStableAreaWriter(inner)
	w.DropBareNewlines = true

	w.WriteRaw([]byte{'\n'})
	if len(inner.ops) != 0 {
		t.Fatalf("lineWrap newline should be dropped, got %v", inner.ops)
	}

	w.ScrollDown()
	w.ScrollUp()
	if got := strings.Join(inner.ops, "|"); got != "SaveCursor|WriteRaw:\n|UnSaveCursor" {
		t.Fatalf("prepareArea newline must still be written, got %q", got)
	}
}

func TestRepeatedPrepareAreaDoesNotAccumulateReserve(t *testing.T) {
	inner := &recorder{}
	w := NewStableAreaWriter(inner)
	for keystroke := 0; keystroke < 8; keystroke++ {
		for i := 0; i < 4; i++ {
			w.ScrollDown()
		}
		for i := 0; i < 4; i++ {
			w.ScrollUp()
		}
		if w.reserved != 0 {
			t.Fatalf("keystroke %d: reserved leaked (%d)", keystroke, w.reserved)
		}
	}
	saves, restores := 0, 0
	for _, op := range inner.ops {
		switch op {
		case "SaveCursor":
			saves++
		case "UnSaveCursor":
			restores++
		}
	}
	if saves != 8 || restores != 8 {
		t.Fatalf("save/restore count saves=%d restores=%d, want 8/8", saves, restores)
	}
}

var _ prompt.ConsoleWriter = (*recorder)(nil)
