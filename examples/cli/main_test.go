package main

import (
	"bytes"
	"strings"
	"testing"
)

// executeCommand runs the root command with the given args and returns stdout output.
func executeCommand(args ...string) (string, error) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestGreetDefault(t *testing.T) {
	out, err := executeCommand("greet", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "World") {
		t.Errorf("greet default output = %q, want Hello World", out)
	}
}

func TestGreetCustom(t *testing.T) {
	out, err := executeCommand("greet", "--name", "Go", "--greeting", "Hi", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Hi") || !strings.Contains(out, "Go") {
		t.Errorf("greet custom output = %q, want Hi Go", out)
	}
}

func TestGreetCount(t *testing.T) {
	out, err := executeCommand("greet", "--name", "World", "--greeting", "Hello", "--count", "3", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(out, "Hello"); count != 3 {
		t.Errorf("greet --count 3 produced %d greetings, want 3", count)
	}
}

func TestListDefault(t *testing.T) {
	out, err := executeCommand("list", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "api-gateway") || !strings.Contains(out, "worker") {
		t.Errorf("list output missing expected items: %q", out)
	}
}

func TestListFilter(t *testing.T) {
	out, err := executeCommand("list", "stopped", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "scheduler") {
		t.Errorf("list filter=stopped should show scheduler: %q", out)
	}
	if strings.Contains(out, "api-gateway") {
		t.Errorf("list filter=stopped should not show api-gateway: %q", out)
	}
}

func TestListJSON(t *testing.T) {
	out, err := executeCommand("list", "--output", "json", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"name"`) || !strings.Contains(out, `"api-gateway"`) {
		t.Errorf("list --output json should produce JSON: %q", out)
	}
}

func TestListNoMatch(t *testing.T) {
	out, err := executeCommand("list", "nonexistent", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "No items match") {
		t.Errorf("list with no matches should warn: %q", out)
	}
}

func TestListTooManyArgs(t *testing.T) {
	_, err := executeCommand("list", "a", "b")
	if err == nil {
		t.Error("list with 2 args should fail")
	}
}

func TestGreetRejectsArgs(t *testing.T) {
	_, err := executeCommand("greet", "unexpected")
	if err == nil {
		t.Error("greet with positional args should fail")
	}
}
