package patterns

import (
	"errors"
	"testing"
	"time"
)

func TestNewServer_Defaults(t *testing.T) {
	s := NewServer("localhost")
	if s.Port != 8080 {
		t.Errorf("Port = %d, want 8080", s.Port)
	}
	if s.MaxConns != 100 {
		t.Errorf("MaxConns = %d, want 100", s.MaxConns)
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	s := NewServer("localhost", WithPort(9090), WithMaxConns(50), WithReadTimeout(3*time.Second))
	if s.Port != 9090 {
		t.Errorf("Port = %d, want 9090", s.Port)
	}
	if s.MaxConns != 50 {
		t.Errorf("MaxConns = %d, want 50", s.MaxConns)
	}
	if s.ReadTimeout != 3*time.Second {
		t.Errorf("ReadTimeout = %v, want 3s", s.ReadTimeout)
	}
}

func TestValidate_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		field   string
	}{
		{"empty", "", true, "name"},
		{"valid", "Alice", false, ""},
		{"too long", string(make([]byte, 51)), true, "name"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if err != nil {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Fatal("expected ValidationError")
				}
				if ve.Field != tc.field {
					t.Errorf("Field = %q, want %q", ve.Field, tc.field)
				}
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	wrapped := Wrap(ErrNotFound, "user lookup")
	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("wrapped error should match ErrNotFound")
	}
	if Wrap(nil, "noop") != nil {
		t.Error("Wrap(nil) should return nil")
	}
}

// TestMultiNotifier_Mock demonstrates mocking via interfaces.
func TestMultiNotifier_Mock(t *testing.T) {
	mock1 := &mockNotifier{}
	mock2 := &mockNotifier{}
	mn := MultiNotifier{mock1, mock2}

	if err := mn.Notify("hello"); err != nil {
		t.Fatal(err)
	}
	if len(mock1.calls) != 1 || mock1.calls[0] != "hello" {
		t.Errorf("mock1 calls = %v", mock1.calls)
	}
	if len(mock2.calls) != 1 {
		t.Errorf("mock2 calls = %v", mock2.calls)
	}
}

func TestMultiNotifier_Error(t *testing.T) {
	fail := &mockNotifier{err: errors.New("fail")}
	mn := MultiNotifier{fail}
	if err := mn.Notify("x"); err == nil {
		t.Error("expected error")
	}
}

// mockNotifier implements Notifier for testing.
type mockNotifier struct {
	calls []string
	err   error
}

func (m *mockNotifier) Notify(msg string) error {
	m.calls = append(m.calls, msg)
	return m.err
}
