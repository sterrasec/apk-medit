package promptio

import "testing"

func TestLiveHelpPanelUnixKeepsOverlay(t *testing.T) {
	env := map[string]string{"TERM": "xterm-256color"}
	if !liveHelpPanel("linux", getenv(env)) {
		t.Fatal("linux + xterm-256color should keep the live help panel")
	}
	if !liveHelpPanel("darwin", getenv(env)) {
		t.Fatal("darwin + xterm-256color should keep the live help panel")
	}
}

func TestLiveHelpPanelDisabledOnWindows(t *testing.T) {
	env := map[string]string{"TERM": "xterm-256color"}
	if liveHelpPanel("windows", getenv(env)) {
		t.Fatal("native Windows should not use the drifting overlay")
	}
}

func TestLiveHelpPanelDisabledOnWindowsHostHints(t *testing.T) {
	cases := []map[string]string{
		{"TERM": "dumb"},
		{"TERM": ""},
		{"TERM": "cygwin"},
		{"WT_SESSION": "abc"},
		{"SESSIONNAME": "Console", "TERM": "xterm-256color"},
		{"OS": "Windows_NT", "TERM": "xterm-256color"},
		{"APK_MEDIT_WINDOWS_CONSOLE": "1", "TERM": "xterm-256color"},
	}
	for i, env := range cases {
		if liveHelpPanel("linux", getenv(env)) {
			t.Fatalf("case %d %v: want live help disabled", i, env)
		}
	}
}

func TestLiveHelpPanelExplicitOverride(t *testing.T) {
	if liveHelpPanel("linux", getenv(map[string]string{
		"TERM":                "xterm-256color",
		"APK_MEDIT_LIVE_HELP": "0",
	})) {
		t.Fatal("APK_MEDIT_LIVE_HELP=0 should disable the overlay")
	}
	if !liveHelpPanel("windows", getenv(map[string]string{
		"APK_MEDIT_LIVE_HELP": "1",
	})) {
		t.Fatal("APK_MEDIT_LIVE_HELP=1 should force the overlay back on")
	}
}

func getenv(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}
