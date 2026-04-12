// Package cli provides reusable helpers for CLI output formatting,
// configuration loading, and common CLI patterns.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// OutputFormat controls how data is rendered.
type OutputFormat string

const (
	// FormatTable renders data as an aligned text table.
	FormatTable OutputFormat = "table"
	// FormatJSON renders data as indented JSON.
	FormatJSON OutputFormat = "json"
	// FormatYAML renders data as YAML.
	FormatYAML OutputFormat = "yaml"
)

// ParseFormat converts a string to OutputFormat, defaulting to table.
func ParseFormat(s string) OutputFormat {
	switch OutputFormat(s) {
	case FormatJSON:
		return FormatJSON
	case FormatYAML:
		return FormatYAML
	default:
		return FormatTable
	}
}

// ANSI color codes for terminal output.
const (
	Reset  = "\033[0m"  // Reset clears all ANSI formatting.
	Red    = "\033[31m" // Red is the ANSI escape for red text.
	Green  = "\033[32m" // Green is the ANSI escape for green text.
	Yellow = "\033[33m" // Yellow is the ANSI escape for yellow text.
	Blue   = "\033[34m" // Blue is the ANSI escape for blue text.
	Bold   = "\033[1m"  // Bold is the ANSI escape for bold text.
)

// Colorize wraps text in ANSI color codes. Returns plain text if noColor is true.
func Colorize(color, text string, noColor bool) string {
	if noColor {
		return text
	}
	return color + text + Reset
}

// Printer handles formatted output to a writer.
type Printer struct {
	// Out is the destination writer (defaults to os.Stdout).
	Out io.Writer
	// Format controls the output rendering (table, JSON, or YAML).
	Format OutputFormat
	// NoColor disables ANSI color codes when true.
	NoColor bool
}

// NewPrinter creates a Printer writing to stdout with the given format.
func NewPrinter(format OutputFormat, noColor bool) *Printer {
	return &Printer{Out: os.Stdout, Format: format, NoColor: noColor}
}

// Print renders data according to the configured format.
// For table format, data should be a [][]string where the first row is headers.
func (p *Printer) Print(data any) error {
	switch p.Format {
	case FormatJSON:
		return p.printJSON(data)
	case FormatYAML:
		return p.printYAML(data)
	default:
		rows, ok := data.([][]string)
		if !ok {
			return fmt.Errorf("table format requires [][]string, got %T", data)
		}
		return p.printTable(rows)
	}
}

func (p *Printer) printJSON(data any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (p *Printer) printYAML(data any) error {
	enc := yaml.NewEncoder(p.Out)
	defer enc.Close()
	return enc.Encode(data)
}

func (p *Printer) printTable(rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}
	w := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	for i, row := range rows {
		for j, col := range row {
			if j > 0 {
				fmt.Fprint(w, "\t")
			}
			if i == 0 {
				fmt.Fprint(w, Colorize(Bold, col, p.NoColor))
			} else {
				fmt.Fprint(w, col)
			}
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

// Success prints a green success message.
func (p *Printer) Success(msg string) {
	fmt.Fprintln(p.Out, Colorize(Green, "✓ "+msg, p.NoColor))
}

// Error prints a red error message.
func (p *Printer) Error(msg string) {
	fmt.Fprintln(p.Out, Colorize(Red, "✗ "+msg, p.NoColor))
}

// Info prints a blue info message.
func (p *Printer) Info(msg string) {
	fmt.Fprintln(p.Out, Colorize(Blue, "ℹ "+msg, p.NoColor))
}

// Warn prints a yellow warning message.
func (p *Printer) Warn(msg string) {
	fmt.Fprintln(p.Out, Colorize(Yellow, "⚠ "+msg, p.NoColor))
}
