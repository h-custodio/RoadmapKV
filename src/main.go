package main

import (
	"RoadmapKV/src/clientserver"
	"RoadmapKV/src/storage"
)

func main() {
	// NewStore() returns pointer to the store
	s := storage.NewStore()
	err := clientserver.CreateServer(s)
	if err != nil {

	}
}
