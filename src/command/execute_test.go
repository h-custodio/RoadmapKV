package command

import (
	"bytes"
	"testing"

	"RoadmapKV/src/storage"
)

func newTestStore() *storage.Store {
	return storage.NewStore()
}

func assertResult(
	t *testing.T,
	got CmdResult,
	status string,
	appendable bool,
) {
	t.Helper()

	if got.Status != status {
		t.Errorf("Status = %q, want %q", got.Status, status)
	}

	if got.IsAppendable != appendable {
		t.Errorf("IsAppendable = %v, want %v", got.IsAppendable, appendable)
	}
}

func TestExecutePing(t *testing.T) {
	got := Execute(newTestStore(), "PING", nil)

	assertResult(t, got, "PONG", false)

	if got.Value != nil {
		t.Errorf("Value = %v, want nil", got.Value)
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	got := Execute(newTestStore(), "UNKNOWN", nil)

	assertResult(t, got, "command unknown", false)

	if got.Value != nil {
		t.Errorf("Value = %v, want nil", got.Value)
	}
}

func TestExecuteSet(t *testing.T) {
	s := newTestStore()

	got := Execute(s, "SET", [][]byte{
		[]byte("name"),
		[]byte("alice"),
	})

	assertResult(t, got, "OK", true)

	if got.Value != nil {
		t.Errorf("Value = %v, want nil", got.Value)
	}

	if value := s.Get("name"); !bytes.Equal(value, []byte("alice")) {
		t.Errorf("stored value = %q, want %q", value, []byte("alice"))
	}
}

func TestExecuteSetInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args [][]byte
	}{
		{name: "no arguments"},
		{name: "one argument", args: [][]byte{[]byte("key")}},
		{
			name: "too many arguments",
			args: [][]byte{
				[]byte("key"),
				[]byte("value"),
				[]byte("extra"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore()

			got := Execute(s, "SET", tt.args)
			assertResult(t, got, "not valid", false)

			if got.Value != nil {
				t.Errorf("Value = %v, want nil", got.Value)
			}

			if len(tt.args) > 0 && s.Exists(string(tt.args[0])) {
				t.Error("invalid SET modified the store")
			}
		})
	}
}

func TestExecuteGet(t *testing.T) {
	s := newTestStore()
	s.Set("name", []byte("alice"))

	got := Execute(s, "GET", [][]byte{[]byte("name")})

	assertResult(t, got, "OK", false)

	if got.Value == nil {
		t.Fatal("GET returned a nil Value pointer")
	}

	if !bytes.Equal(*got.Value, []byte("alice")) {
		t.Errorf("GET value = %q, want %q", *got.Value, []byte("alice"))
	}
}

func TestExecuteGetMissingKey(t *testing.T) {
	got := Execute(newTestStore(), "GET", [][]byte{[]byte("missing")})

	assertResult(t, got, "OK", false)

	if got.Value == nil {
		t.Fatal("GET should return a Value pointer, even for a missing key")
	}

	if *got.Value != nil {
		t.Errorf("GET missing key = %v, want nil", *got.Value)
	}
}

func TestExecuteDelete(t *testing.T) {
	s := newTestStore()
	s.Set("name", []byte("alice"))

	got := Execute(s, "DEL", [][]byte{[]byte("name")})

	assertResult(t, got, "OK", true)

	if got.Value != nil {
		t.Errorf("Value = %v, want nil", got.Value)
	}

	if s.Exists("name") {
		t.Error("key still exists after DEL")
	}
}

func TestExecuteDeleteMissingKey(t *testing.T) {
	s := newTestStore()

	got := Execute(s, "DEL", [][]byte{[]byte("missing")})

	assertResult(t, got, "OK", true)

	if s.Exists("missing") {
		t.Error("DEL created a previously missing key")
	}
}

func TestExecuteSetGetDeleteIntegration(t *testing.T) {
	s := newTestStore()

	setResult := Execute(s, "SET", [][]byte{
		[]byte("color"),
		[]byte("blue"),
	})
	assertResult(t, setResult, "OK", true)

	getResult := Execute(s, "GET", [][]byte{[]byte("color")})
	assertResult(t, getResult, "OK", false)

	if getResult.Value == nil {
		t.Fatal("GET returned a nil Value pointer")
	}
	if !bytes.Equal(*getResult.Value, []byte("blue")) {
		t.Fatalf("GET value = %q, want %q", *getResult.Value, []byte("blue"))
	}

	delResult := Execute(s, "DEL", [][]byte{[]byte("color")})
	assertResult(t, delResult, "OK", true)

	getResult = Execute(s, "GET", [][]byte{[]byte("color")})
	if getResult.Value == nil {
		t.Fatal("GET returned a nil Value pointer after deletion")
	}
	if *getResult.Value != nil {
		t.Errorf("GET after DEL = %v, want nil", *getResult.Value)
	}
}

func TestExecuteGetAndDeleteInvalidArguments(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    [][]byte
	}{
		{name: "GET no arguments", command: "GET"},
		{
			name:    "GET too many arguments",
			command: "GET",
			args:    [][]byte{[]byte("key"), []byte("extra")},
		},
		{name: "DEL no arguments", command: "DEL"},
		{
			name:    "DEL too many arguments",
			command: "DEL",
			args:    [][]byte{[]byte("key"), []byte("extra")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore()
			s.Set("key", []byte("original"))

			// A malformed command should return an error result,
			// not panic or modify existing data.
			var got CmdResult
			panicked := false

			func() {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				got = Execute(s, tt.command, tt.args)
			}()

			if panicked {
				t.Fatal("Execute panicked on invalid arguments")
			}

			assertResult(t, got, "not valid", false)

			if value := s.Get("key"); !bytes.Equal(value, []byte("original")) {
				t.Errorf("invalid command changed stored value to %q", value)
			}
		})
	}
}
