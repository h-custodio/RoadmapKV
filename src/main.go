package main

import (
	"RoadmapKV/src/clientserver"
	"RoadmapKV/src/storage"
)

func main() {
	// NewStore() returns pointer to the store
	s := storage.NewStore()
	clientserver.CreateServer(s)
}
