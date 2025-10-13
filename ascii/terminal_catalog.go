package ascii

import (
	"os"
	"strings"
)

type TerminalType int

const (
	TerminalUnknown TerminalType = iota
	TerminalXterm
	TerminalITerm
	TerminalGhostty
	TerminalWindowsTerminal
	TerminalGeneric
)

func DetectTerminal() TerminalType {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	switch {
	case strings.Contains(termProgram, "ghostty") || strings.Contains(term, "ghostty"):
		return TerminalGhostty
	case strings.Contains(termProgram, "iTerm"):
		return TerminalITerm
	case strings.Contains(termProgram, "WindowsTerminal"):
		return TerminalWindowsTerminal
	case strings.Contains(term, "xterm"):
		return TerminalXterm
	default:
		return TerminalGeneric
	}
}

func GetTerminalOverrides(termType TerminalType) map[rune]int {
	return make(map[rune]int)
}
