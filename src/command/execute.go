package command

import (
	"RoadmapKV/src/storage"
	"os"
)

type CmdResultResult struct {
	Status       string
	IsAppendable bool
	Value        *[]byte
}

func Execute(s *storage.Store, command string, args [][]byte) CmdResultResult {
	switch command {
	case "SET":
		return set(s, args)

	case "GET":
		return get(s, string(args[0]))

	case "DEL":
		return del(s, string(args[0]))

	case "PING":
		return CmdResultResult{
			Status:       "PONG",
			IsAppendable: false,
			Value:        nil,
		}

	case "QUIT":
		return quit()

	default:
		return CmdResultResult{
			Status:       "command unknown",
			IsAppendable: false,
			Value:        nil,
		}
	}
}

func set(s *storage.Store, args [][]byte) CmdResultResult {
	if len(args) != 2 {
		return CmdResultResult{
			Status:       "not valid",
			IsAppendable: false,
			Value:        nil,
		}
	}

	s.Set(string(args[0]), args[1])
	return CmdResultResult{
		Status:       "OK",
		IsAppendable: true,
		Value:        nil,
	}
}

func get(s *storage.Store, key string) CmdResultResult {
	value := s.Get(key)
	return CmdResultResult{
		Status:       "OK",
		IsAppendable: false,
		Value:        &value,
	}
}

func del(s *storage.Store, key string) CmdResultResult {
	s.Delete(key)
	return CmdResultResult{
		Status:       "OK",
		IsAppendable: true,
		Value:        nil,
	}
}

func quit() CmdResultResult {
	os.Exit(0)
	return CmdResultResult{
		Status:       "Quiting...",
		IsAppendable: false,
		Value:        nil,
	}
}
