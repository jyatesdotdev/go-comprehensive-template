package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  OutputFormat
	}{
		{"json", FormatJSON},
		{"yaml", FormatYAML},
		{"table", FormatTable},
		{"", FormatTable},
		{"unknown", FormatTable},
	}
	for _, tt := range tests {
		if got := ParseFormat(tt.input); got != tt.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestColorize(t *testing.T) {
	got := Colorize(Red, "error", false)
	if !strings.Contains(got, Red) || !strings.HasSuffix(got, Reset) {
		t.Errorf("Colorize with color should wrap text, got %q", got)
	}
	if got := Colorize(Red, "error", true); got != "error" {
		t.Errorf("Colorize noColor should return plain text, got %q", got)
	}
}

func newTestPrinter(format OutputFormat, noColor bool) (*Printer, *bytes.Buffer) {
	var buf bytes.Buffer
	return &Printer{Out: &buf, Format: format, NoColor: noColor}, &buf
}

func TestPrinterJSON(t *testing.T) {
	p, buf := newTestPrinter(FormatJSON, true)
	data := map[string]string{"key": "value"}
	if err := p.Print(data); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("got %v, want key=value", got)
	}
}

func TestPrinterYAML(t *testing.T) {
	p, buf := newTestPrinter(FormatYAML, true)
	data := map[string]string{"key": "value"}
	if err := p.Print(data); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("got %v, want key=value", got)
	}
}

func TestPrinterTable(t *testing.T) {
	p, buf := newTestPrinter(FormatTable, true)
	rows := [][]string{{"NAME", "STATUS"}, {"api", "running"}}
	if err := p.Print(rows); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "api") {
		t.Errorf("table output missing expected content: %q", out)
	}
}

func TestPrinterTableEmpty(t *testing.T) {
	p, buf := newTestPrinter(FormatTable, true)
	if err := p.Print([][]string{}); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("empty table should produce no output, got %q", buf.String())
	}
}

func TestPrinterTableWrongType(t *testing.T) {
	p, _ := newTestPrinter(FormatTable, true)
	if err := p.Print("not a table"); err == nil {
		t.Error("expected error for non-[][]string data in table format")
	}
}

func TestPrinterStatusMessages(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(*Printer, string)
		prefix string
	}{
		{"Success", (*Printer).Success, "✓"},
		{"Error", (*Printer).Error, "✗"},
		{"Info", (*Printer).Info, "ℹ"},
		{"Warn", (*Printer).Warn, "⚠"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, buf := newTestPrinter(FormatTable, true)
			tt.fn(p, "test message")
			out := buf.String()
			if !strings.Contains(out, tt.prefix) || !strings.Contains(out, "test message") {
				t.Errorf("%s output = %q, want prefix %q and message", tt.name, out, tt.prefix)
			}
		})
	}
}

func TestPrinterStatusMessagesWithColor(t *testing.T) {
	p, buf := newTestPrinter(FormatTable, false)
	p.Success("ok")
	if !strings.Contains(buf.String(), Green) {
		t.Error("Success with color should contain ANSI green code")
	}
}
