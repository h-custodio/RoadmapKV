package command

import (
	"RoadmapKV/src/storage"
)

type CmdResult struct {
	Response     *[]byte
	IsAppendable bool
	CloseConn    bool
}

func Execute(s *storage.Store, command string, args [][]byte) CmdResult {
	switch command {
	case "SET":
		return set(s, args)

	case "GET":
		return get(s, string(args[0]))

	case "DEL":
		return del(s, string(args[0]))

	case "PING":
		return CmdResult{
			Response:     "PONG",
			IsAppendable: false,
		}

	case "QUIT":
		return quit()

	default:
		return CmdResult{
			Response: "command unknown",
		}
	}
}

func set(s *storage.Store, args [][]byte) CmdResult {
	if len(args) != 2 {
		return CmdResult{
			Response:     "not valid",
			IsAppendable: false,
		}
	}

	s.Set(string(args[0]), args[1])
	return CmdResult{
		Response:     "OK",
		IsAppendable: true,
	}
}

func get(s *storage.Store, key string) CmdResult {
	value := s.Get(key)
	return CmdResult{
		Response:     &value,
		IsAppendable: false,
	}
}

func del(s *storage.Store, key string) CmdResult {
	s.Delete(key)
	return CmdResult{
		Response:     "OK",
		IsAppendable: true,
	}
}

func quit() CmdResult {
	return CmdResult{
		Response:  "Quiting...",
		CloseConn: true,
	}
}
