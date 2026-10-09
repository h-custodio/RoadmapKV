package command

import (
	"RoadmapKV/src/storage"
	"os"
)

type Result struct {
	status       string
	isAppendable bool
	value        *[]byte
}

func Execute(s *storage.Store, command string, args [][]byte) Result {
	switch command {
	case "SET":
		return set(s, args)

	case "GET":
		return get(s, string(args[0]))

	case "DEL":
		return del(s, string(args[0]))

	case "PING":
		return Result{
			status:       "PONG",
			isAppendable: false,
			value:        nil,
		}

	case "QUIT":
		return quit()

	default:
		return Result{
			status:       "command unknown",
			isAppendable: false,
			value:        nil,
		}
	}
}

func set(s *storage.Store, args [][]byte) Result {
	if len(args) != 2 {
		return Result{
			status:       "not valid",
			isAppendable: false,
			value:        nil,
		}
	}

	s.Set(string(args[0]), args[1])
	return Result{
		status:       "OK",
		isAppendable: true,
		value:        nil,
	}
}

func get(s *storage.Store, key string) Result {
	val := s.Get(key)
	return Result{
		status:       "OK",
		isAppendable: false,
		value:        &val,
	}
}

func del(s *storage.Store, key string) Result {
	s.Delete(key)
	return Result{
		status:       "OK",
		isAppendable: true,
		value:        nil,
	}
}

func quit() Result {
	os.Exit(0)
	return Result{
		status:       "Quiting...",
		isAppendable: false,
		value:        nil,
	}
}
