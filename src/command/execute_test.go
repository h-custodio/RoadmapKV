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
	response []byte,
	appendable bool,
) {
	t.Helper()

	if !bytes.Equal(got.Response, response) {
		t.Errorf("Response = %q, want %q", got.Response, response)
	}

	if got.IsAppendable != appendable {
		t.Errorf("IsAppendable = %v, want %v", got.IsAppendable, appendable)
	}
}

func TestExecutePing(t *testing.T) {
	got := Execute(newTestStore(), "PING", nil)

	assertResult(t, got, []byte("PONG"), false)
}

func TestExecuteUnknownCommand(t *testing.T) {
	got := Execute(newTestStore(), "UNKNOWN", nil)

	assertResult(t, got, []byte("command unknown"), false)
}

func TestExecuteSet(t *testing.T) {
	s := newTestStore()

	got := Execute(s, "SET", [][]byte{
		[]byte("name"),
		[]byte("alice"),
	})

	assertResult(t, got, []byte("OK"), true)

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
			assertResult(t, got, []byte("not valid"), false)

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

	assertResult(t, got, []byte("alice"), false)
}

func TestExecuteGetMissingKey(t *testing.T) {
	got := Execute(newTestStore(), "GET", [][]byte{[]byte("missing")})

	assertResult(t, got, nil, false)
}

func TestExecuteDelete(t *testing.T) {
	s := newTestStore()
	s.Set("name", []byte("alice"))

	got := Execute(s, "DEL", [][]byte{[]byte("name")})

	assertResult(t, got, []byte("OK"), true)

	if s.Exists("name") {
		t.Error("key still exists after DEL")
	}
}

func TestExecuteDeleteMissingKey(t *testing.T) {
	s := newTestStore()

	got := Execute(s, "DEL", [][]byte{[]byte("missing")})

	assertResult(t, got, []byte("OK"), true)

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
	assertResult(t, setResult, []byte("OK"), true)

	getResult := Execute(s, "GET", [][]byte{[]byte("color")})
	assertResult(t, getResult, []byte("blue"), false)

	delResult := Execute(s, "DEL", [][]byte{[]byte("color")})
	assertResult(t, delResult, []byte("OK"), true)

	getResult = Execute(s, "GET", [][]byte{[]byte("color")})
	if getResult.Response != nil {
		t.Errorf("GET after DEL = %v, want nil", getResult.Response)
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

			assertResult(t, got, []byte("not valid"), false)

			if value := s.Get("key"); !bytes.Equal(value, []byte("original")) {
				t.Errorf("invalid command changed stored value to %q", value)
			}
		})
	}
}
