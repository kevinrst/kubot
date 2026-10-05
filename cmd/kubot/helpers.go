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
