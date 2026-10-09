package main

import (
	"RoadmapKV/src/clientserver"
	"RoadmapKV/src/storage"
)

func main() {
	s := storage.Store{}
	clientserver.CreateServer(&s)
}
