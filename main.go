package main

import (
	"github.com/memsql/online-upgrade/util"
	"log"
)

func main() {
	util.SetupLogging()
	config := util.ParseFlags()
	if err := util.ConnectToMemSQL(config); err != nil {
		log.Fatalf("Failed to connect to MemSQL: %s", err)
	}
}
