package main

import (
	"log"

	"github.com/sjzar/file-store-mcp/cmd/filestore"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	filestore.Execute()
}
