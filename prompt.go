package main

import (
	"fmt"
	"os"

	"github.com/sterrasec/apk-medit/pkg/promptio"

	prompt "github.com/c-bata/go-prompt"
)

var commandHelp = []prompt.Suggest{
	{Text: "find   <int>", Description: "Search the specified integer."},
	{Text: "find   <datatype> <int>", Description: "Types can be specified are string, word, dword, qword."},
	{Text: "filter <int>", Description: "Filter previous search results that match the current search results."},
	{Text: "patch  <int>", Description: "Write the specified value on the address found by search."},
	{Text: "attach", Description: "Attach to the target process by ptrace."},
	{Text: "detach", Description: "Detach from the attached process."},
	{Text: "ps", Description: "Find the target process and if there is only one, specify it as the target."},
	{Text: "dump <begin addr> <end addr>", Description: "Display memory dump like hexdump"},
	{Text: "exit"},
}

var commandNames = []prompt.Suggest{
	{Text: "find"},
	{Text: "filter"},
	{Text: "patch"},
	{Text: "attach"},
	{Text: "detach"},
	{Text: "ps"},
	{Text: "dump"},
	{Text: "exit"},
}

func completer(t prompt.Document) []prompt.Suggest {
	return complete(t, promptio.LiveHelpPanel())
}

func complete(t prompt.Document, liveHelp bool) []prompt.Suggest {
	if liveHelp {
		return commandHelp
	}
	word := t.GetWordBeforeCursor()
	if word == "" {
		return nil
	}
	return prompt.FilterHasPrefix(commandNames, word, true)
}

func newPromptWriter() prompt.ConsoleWriter {
	w := promptio.NewStableAreaWriter(prompt.NewStdoutWriter())
	if !promptio.LiveHelpPanel() {
		w.DropBareNewlines = true
	}
	return w
}

func printCommandHelp(out *os.File) {
	fmt.Fprintln(out, "Commands:")
	for _, s := range commandHelp {
		if s.Description == "" {
			fmt.Fprintf(out, "  %s\n", s.Text)
			continue
		}
		fmt.Fprintf(out, "  %-28s %s\n", s.Text, s.Description)
	}
	fmt.Fprintln(out, "Live command descriptions are disabled on this console so the input line stays visible.")
	fmt.Fprintln(out, "Tab-complete command names, or set APK_MEDIT_LIVE_HELP=1 to restore the overlay.")
}

func promptOptions() []prompt.Option {
	opts := []prompt.Option{
		prompt.OptionTitle("medit: MEmory eDIT tool"),
		prompt.OptionPrefix("> "),
		prompt.OptionInputTextColor(prompt.Cyan),
		prompt.OptionPrefixTextColor(prompt.DarkBlue),
		prompt.OptionPreviewSuggestionTextColor(prompt.Green),
		prompt.OptionDescriptionTextColor(prompt.DarkGray),
		prompt.OptionWriter(newPromptWriter()),
	}
	return opts
}
