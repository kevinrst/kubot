package main

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

func argAt(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Reports whether stdin is a terminal (skip prompts when piped).
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Honors --no-color, NO_COLOR, and non-TTY stdout.
func useColor(noColorFlag bool) bool {
	if noColorFlag || os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Stdout width, or 100 when piped.
func terminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if c := os.Getenv("COLUMNS"); c != "" {
		if w, err := strconv.Atoi(c); err == nil && w > 0 {
			return w
		}
	}
	return 100
}
