package command

import (
	"RoadmapKV/src/storage"
	"strings"
)

type CmdResult struct {
	Response     []byte
	IsAppendable bool
	CloseConn    bool
}

func Execute(s *storage.Store, args [][]byte) CmdResult {
	if len(args) == 0 {
		return CmdResult{
			Response: []byte("empty command"),
		}
	}

	command := strings.ToUpper(string(args[0]))

	switch command {
	case "SET":
		return set(s, args)

	case "GET":
		return get(s, args)

	case "DEL":
		return del(s, args)

	case "PING":
		return CmdResult{
			Response:     []byte("PONG"),
			IsAppendable: false,
		}

	case "QUIT":
		return quit()

	default:
		return CmdResult{
			Response: []byte("command unknown"),
		}
	}
}

func set(s *storage.Store, args [][]byte) CmdResult {
	if len(args) != 3 {
		return CmdResult{
			Response:     []byte("not valid"),
			IsAppendable: false,
		}
	}

	// args[0] is set, args[1] is key, args[2] is value
	s.Set(string(args[1]), args[2])
	return CmdResult{
		Response:     []byte("OK"),
		IsAppendable: true,
	}
}

func get(s *storage.Store, args [][]byte) CmdResult {
	if len(args) != 2 {
		return CmdResult{
			Response:     []byte("not valid"),
			IsAppendable: false,
		}
	}

	key := args[1]
	value := s.Get(string(key))

	if value == nil {
		return CmdResult{
			Response:     []byte("not valid"),
			IsAppendable: false,
		}
	}

	return CmdResult{
		Response:     value,
		IsAppendable: false,
	}
}

func del(s *storage.Store, args [][]byte) CmdResult {
	if len(args) != 2 {
		return CmdResult{
			Response:     []byte("not valid"),
			IsAppendable: false,
		}
	}

	key := args[1]
	success := s.Delete(string(key))

	if !success {
		return CmdResult{
			Response:     []byte("not valid"),
			IsAppendable: false,
		}
	}

	return CmdResult{
		Response:     []byte("OK"),
		IsAppendable: true,
	}
}

func quit() CmdResult {
	return CmdResult{
		Response:  []byte("Quitting..."),
		CloseConn: true,
	}
}
