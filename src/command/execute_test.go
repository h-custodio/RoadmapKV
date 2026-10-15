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
	got := Execute(newTestStore(), [][]byte{[]byte("PING")})

	assertResult(t, got, []byte("PONG\n"), false)
}

func TestExecuteUnknownCommand(t *testing.T) {
	got := Execute(newTestStore(), [][]byte{[]byte("UNKNOWN")})

	assertResult(t, got, []byte("command unknown\n"), false)
}

func TestExecuteSet(t *testing.T) {
	s := newTestStore()

	got := Execute(s, [][]byte{
		[]byte("SET"),
		[]byte("name"),
		[]byte("alice"),
	})

	assertResult(t, got, []byte("OK\n"), true)

	if value := s.Get("name"); !bytes.Equal(value, []byte("alice")) {
		t.Errorf("stored value = %q, want %q", value, []byte("alice"))
	}
}

func TestExecuteSetInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args [][]byte
	}{
		{name: "no arguments", args: [][]byte{}},
		{name: "one argument", args: [][]byte{[]byte("SET"), []byte("key")}},
		{
			name: "too many arguments",
			args: [][]byte{
				[]byte("SET"),
				[]byte("key"),
				[]byte("value"),
				[]byte("extra"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore()

			got := Execute(s, tt.args)
			// Empty args returns "empty command", malformed SET returns "not valid".
			expectedResp := []byte("not valid\n")
			if len(tt.args) == 0 {
				expectedResp = []byte("empty command\n")
			}
			assertResult(t, got, expectedResp, false)

			if len(tt.args) > 1 && s.Exists(string(tt.args[1])) {
				t.Error("invalid SET modified the store")
			}
		})
	}
}

func TestExecuteGet(t *testing.T) {
	s := newTestStore()
	s.Set("name", []byte("alice"))

	got := Execute(s, [][]byte{[]byte("GET"), []byte("name")})

	assertResult(t, got, []byte("alice"), false)
}

func TestExecuteGetMissingKey(t *testing.T) {
	got := Execute(newTestStore(), [][]byte{[]byte("GET"), []byte("missing")})

	// Updated implementation returns "not valid" for missing keys on GET
	assertResult(t, got, []byte("not valid\n"), false)
}

func TestExecuteDelete(t *testing.T) {
	s := newTestStore()
	s.Set("name", []byte("alice"))

	got := Execute(s, [][]byte{[]byte("DEL"), []byte("name")})

	assertResult(t, got, []byte("OK\n"), true)

	if s.Exists("name") {
		t.Error("key still exists after DEL")
	}
}

func TestExecuteDeleteMissingKey(t *testing.T) {
	s := newTestStore()

	got := Execute(s, [][]byte{[]byte("DEL"), []byte("missing")})

	// Updated implementation returns "not valid" if delete fails (missing key)
	assertResult(t, got, []byte("not valid\n"), false)
}

func TestExecuteSetGetDeleteIntegration(t *testing.T) {
	s := newTestStore()

	setResult := Execute(s, [][]byte{
		[]byte("SET"),
		[]byte("color"),
		[]byte("blue"),
	})
	assertResult(t, setResult, []byte("OK\n"), true)

	getResult := Execute(s, [][]byte{[]byte("GET"), []byte("color")})
	assertResult(t, getResult, []byte("blue"), false)

	delResult := Execute(s, [][]byte{[]byte("DEL"), []byte("color")})
	assertResult(t, delResult, []byte("OK\n"), true)

	getResult = Execute(s, [][]byte{[]byte("GET"), []byte("color")})
	assertResult(t, getResult, []byte("not valid\n"), false)
}

func TestExecuteGetAndDeleteInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args [][]byte
	}{
		{name: "GET no arguments", args: [][]byte{[]byte("GET")}},
		{
			name: "GET too many arguments",
			args: [][]byte{[]byte("GET"), []byte("key"), []byte("extra")},
		},
		{name: "DEL no arguments", args: [][]byte{[]byte("DEL")}},
		{
			name: "DEL too many arguments",
			args: [][]byte{[]byte("DEL"), []byte("key"), []byte("extra")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore()
			s.Set("key", []byte("original"))

			var got CmdResult
			panicked := false

			func() {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()

				got = Execute(s, tt.args)
			}()

			if panicked {
				t.Fatal("Execute panicked on invalid arguments")
			}

			assertResult(t, got, []byte("not valid\n"), false)

			if value := s.Get("key"); !bytes.Equal(value, []byte("original")) {
				t.Errorf("invalid command changed stored value to %q", value)
			}
		})
	}
}
