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
func LiveHelpPanel() bool {
	return liveHelpPanel(runtime.GOOS, os.Getenv)
}

func liveHelpPanel(goos string, getenv func(string) string) bool {
	if explicit := strings.TrimSpace(getenv("APK_MEDIT_LIVE_HELP")); explicit != "" {
		switch strings.ToLower(explicit) {
		case "0", "false", "off", "no":
			return false
		default:
			return true
		}
	}
	if windowsConsoleCompat(goos, getenv) {
		return false
	}
	return true
}

// windowsConsoleCompat is true when the process is a native Windows
// binary or the host console is a Windows-like conhost that mishandles
// go-prompt's suggestion rendering.
func windowsConsoleCompat(goos string, getenv func(string) string) bool {
	if goos == "windows" {
		return true
	}
	if v := strings.TrimSpace(getenv("APK_MEDIT_WINDOWS_CONSOLE")); v != "" {
		switch strings.ToLower(v) {
		case "0", "false", "off", "no":
			return false
		default:
			return true
		}
	}
	if getenv("WT_SESSION") != "" {
		return true
	}
	if strings.EqualFold(getenv("SESSIONNAME"), "Console") {
		return true
	}
	if strings.EqualFold(getenv("OS"), "Windows_NT") {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(getenv("TERM"))) {
	case "", "dumb", "cygwin", "win32", "win32con":
		return true
	}
	return false
}
