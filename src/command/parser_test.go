package command

import (
	"bytes"
	"reflect"
	"testing"
)

func assertLine(t *testing.T, got, want [][]byte) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Line = %q, want %q", got, want)
	}
}

func TestParseBasicCommand(t *testing.T) {
	got := Parse([]byte("SET name alice\n"))

	assertLine(t, got.Line, [][]byte{
		[]byte("SET"),
		[]byte("name"),
		[]byte("alice"),
	})

	if !got.IsComplete {
		t.Error("IsComplete = false, want true")
	}

	if got.BytesConsumed != len("SET name alice\n") {
		t.Errorf("BytesConsumed = %d, want %d",
			got.BytesConsumed, len("SET name alice\n"))
	}

	if got.InPound {
		t.Error("InPound = true after complete command, want false")
	}
}

func TestParseQuotedValue(t *testing.T) {
	got := Parse([]byte("SET message #hello world#\n"))

	assertLine(t, got.Line, [][]byte{
		[]byte("SET"),
		[]byte("message"),
		[]byte("hello world"),
	})

	if !got.IsComplete {
		t.Error("quoted command should be complete")
	}
}

func TestParseMultipleCommands(t *testing.T) {
	data := []byte("PING\nSET key value\n")

	first := Parse(data)

	assertLine(t, first.Line, [][]byte{
		[]byte("PING"),
	})

	if !first.IsComplete {
		t.Fatal("first command should be complete")
	}

	if first.BytesConsumed != len("PING\n") {
		t.Fatalf("first BytesConsumed = %d, want %d",
			first.BytesConsumed, len("PING\n"))
	}

	second := Parse(data[first.BytesConsumed:])

	assertLine(t, second.Line, [][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte("value"),
	})

	if !second.IsComplete {
		t.Error("second command should be complete")
	}
}

func TestParseIncompleteCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		inPound bool
	}{
		{
			name:  "unfinished argument",
			input: "SET key value",
		},
		{
			name:    "unfinished quoted value",
			input:   "SET key #hello world",
			inPound: true,
		},
		{
			name:  "empty input",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse([]byte(tt.input))

			if got.IsComplete {
				t.Error("IsComplete = true, want false")
			}

			if got.BytesConsumed != len(tt.input) {
				t.Errorf("BytesConsumed = %d, want %d",
					got.BytesConsumed, len(tt.input))
			}

			if got.InPound != tt.inPound {
				t.Errorf("InPound = %v, want %v",
					got.InPound, tt.inPound)
			}
		})
	}
}

func TestParsePartialInputThenComplete(t *testing.T) {
	// The caller retains incomplete input and reparses the entire
	// accumulated buffer when more bytes arrive.
	partial := []byte("SET key #hello")
	first := Parse(partial)

	if first.IsComplete {
		t.Fatal("partial input should not be complete")
	}

	combined := append(append([]byte(nil), partial...), []byte(" world#\n")...)
	got := Parse(combined)

	assertLine(t, got.Line, [][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte("hello world"),
	})

	if !got.IsComplete {
		t.Error("command should be complete after receiving remaining bytes")
	}

	if got.BytesConsumed != len(combined) {
		t.Errorf("BytesConsumed = %d, want %d",
			got.BytesConsumed, len(combined))
	}
}

func TestParseDoesNotConsumeBytesFromNextCommand(t *testing.T) {
	data := []byte("PING\nSET key value\n")

	got := Parse(data)

	if got.BytesConsumed != len("PING\n") {
		t.Fatalf("BytesConsumed = %d, want %d",
			got.BytesConsumed, len("PING\n"))
	}

	remaining := data[got.BytesConsumed:]
	if !bytes.Equal(remaining, []byte("SET key value\n")) {
		t.Errorf("remaining input = %q, want %q",
			remaining, "SET key value\n")
	}
}

func TestParseEmptyValue(t *testing.T) {
	got := Parse([]byte("SET key ##\n"))

	assertLine(t, got.Line, [][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte(""),
	})

	if !got.IsComplete {
		t.Error("empty quoted value should be accepted as a complete command")
	}
}

func TestParsePreservesUTF8Bytes(t *testing.T) {
	input := []byte("SET greeting café\n")

	got := Parse(input)

	assertLine(t, got.Line, [][]byte{
		[]byte("SET"),
		[]byte("greeting"),
		[]byte("café"),
	})

	if got.BytesConsumed != len(input) {
		t.Errorf("BytesConsumed = %d, want %d",
			got.BytesConsumed, len(input))
	}
}

func TestParsePreservesArbitraryValueBytes(t *testing.T) {
	value := []byte{0x00, 0x7f, 0x80, 0xff}
	input := append([]byte("SET key "), value...)
	input = append(input, '\n')

	got := Parse(input)

	if !got.IsComplete {
		t.Fatal("command should be complete")
	}

	if len(got.Line) != 3 {
		t.Fatalf("len(Line) = %d, want 3", len(got.Line))
	}

	if !bytes.Equal(got.Line[2], value) {
		t.Errorf("parsed value = %v, want %v", got.Line[2], value)
	}

	if got.BytesConsumed != len(input) {
		t.Errorf("BytesConsumed = %d, want %d",
			got.BytesConsumed, len(input))
	}
}
