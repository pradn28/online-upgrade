package main

import (
	"fmt"
	"github.com/memsql/online-upgrade/util"
	"log"
)

func main() {
	config := util.ParseFlags()

	logFinalizer, err := util.SetupLogging(config)
	if err != nil {
		panic(fmt.Errorf("Failed to setup logging: %s", err))
	}
	defer logFinalizer()

	if err := util.ConnectToMemSQL(config); err != nil {
		log.Fatalf("Failed to connect to MemSQL: %s", err)
	}
}
