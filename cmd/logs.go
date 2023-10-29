package cmd

import (
	"fmt"
	"os"

	"github.com/gookit/color"
)

// Log prints if verbose is enabled, and adds some color if it is enabled.
func (s *config) Log(format string, a ...interface{}) {

	if !s.verbose {
		return
	}
	format = "> " + format
	if s.color {
		format = color.C256(214).Sprint(format)
	}
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

// Err prints an error message, adds some color if enabled.
func (s *config) Err(format string, a ...interface{}) {

	if s.color {
		format = color.Red.Sprint(format)
	}
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

// Info prints a message to stderr that won't be picked up by a standard simple
// pipe/redirection.
func (s *config) Info(format string, a ...interface{}) {

	if s.color {
		format = color.Blue.Sprint(format)
	}
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func Newline() {
	_, _ = fmt.Fprintf(os.Stderr, "\n")
}

// Fatal prints an error, and then terminates the program.
func (s *config) Fatal(format string, a ...interface{}) {

	format = "\nFATAL: " + format

	if s.color {
		format = color.Red.Sprint(format)
	}
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
