package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	prompt "github.com/c-bata/go-prompt"
)

func TestCompleteLiveHelpReturnsAllCommands(t *testing.T) {
	got := complete(prompt.Document{}, true)
	if len(got) != len(commandHelp) {
		t.Fatalf("live help should return all %d commands, got %d", len(commandHelp), len(got))
	}
}

func TestCompleteStableConsoleHidesEmptyOverlay(t *testing.T) {
	got := complete(prompt.Document{}, false)
	if len(got) != 0 {
		t.Fatalf("empty input must not open a help overlay on Windows-like consoles, got %v", got)
	}
}

func TestCompleteStableConsoleFiltersCommandNames(t *testing.T) {
	got := complete(documentAtEnd("fi"), false)
	if len(got) != 2 {
		t.Fatalf("typed fi: want find and filter, got %v", got)
	}
	if got[0].Text != "find" || got[1].Text != "filter" {
		t.Fatalf("unexpected suggestions %v", got)
	}
	for _, s := range got {
		if s.Description != "" {
			t.Fatalf("Windows-like completions must stay narrow (no descriptions), got %q", s.Description)
		}
	}

	got = complete(documentAtEnd("dump"), false)
	if len(got) != 1 || got[0].Text != "dump" {
		t.Fatalf("typed dump: want dump, got %v", got)
	}

	got = complete(documentAtEnd("zzz"), false)
	if len(got) != 0 {
		t.Fatalf("no match should hide the panel, got %v", got)
	}
}

func TestPrintCommandHelpListsFindAndPatch(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	printCommandHelp(w)
	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"find", "patch", "APK_MEDIT_LIVE_HELP"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func documentAtEnd(text string) prompt.Document {
	buf := prompt.NewBuffer()
	buf.InsertText(text, false, true)
	return *buf.Document()
}
