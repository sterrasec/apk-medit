package promptio

import (
	"os"
	"runtime"
	"strings"
)

// LiveHelpPanel reports whether go-prompt should keep the always-on
// command-description overlay.
//
// That overlay is re-rendered on every keystroke. On Windows consoles
// (native builds, and Linux/Android binaries viewed from W10 Cmd or
// PowerShell via adb) go-prompt's cursor math and IND/RI scrolling do
// not stay in sync, so the panel walks upward and covers the input line.
// See https://github.com/sterrasec/apk-medit/issues/37.
//
// APK_MEDIT_LIVE_HELP=1/0 forces the overlay on or off regardless of the
// detected console.
func LiveHelpPanel() bool {
	return liveHelpPanel(runtime.GOOS, os.Getenv)
}

func liveHelpPanel(goos string, getenv func(string) string) bool {
	if explicit := strings.TrimSpace(getenv("APK_MEDIT_LIVE_HELP")); explicit != "" {
		return !isFalse(explicit)
	}
	return !windowsConsoleCompat(goos, getenv)
}

// windowsConsoleCompat is true when the process is a native Windows
// binary or the terminal looks like a Windows-like console that
// mishandles go-prompt's suggestion rendering.
//
// Only TERM is consulted because it is the one host variable adb forwards
// into the device shell. Host-side hints such as WT_SESSION or
// OS=Windows_NT never reach an Android process, and WT_SESSION also leaks
// into WSL where xterm rendering works fine. APK_MEDIT_WINDOWS_CONSOLE=1/0
// overrides the TERM heuristic.
func windowsConsoleCompat(goos string, getenv func(string) string) bool {
	if goos == "windows" {
		return true
	}
	if v := strings.TrimSpace(getenv("APK_MEDIT_WINDOWS_CONSOLE")); v != "" {
		return !isFalse(v)
	}
	switch strings.ToLower(strings.TrimSpace(getenv("TERM"))) {
	case "", "dumb", "cygwin", "win32", "win32con":
		return true
	}
	return false
}

func isFalse(v string) bool {
	switch strings.ToLower(v) {
	case "0", "false", "off", "no":
		return true
	}
	return false
}
