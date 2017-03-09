package main

import (
	"fmt"
	"log"

	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/util"
)

func main() {
	// Grab any config information specified on the command line
	// If no config information is passed in, we will use the defaults in 'config.go'
	config := util.ParseFlags()

	// Open log file
	logFinalizer, err := util.SetupLogging(config)
	if err != nil {
		panic(fmt.Errorf("Failed to setup logging: %s", err))
	}

	// Defer closing the log file until upgrade is complete
	defer logFinalizer()

	// Connect to MemSQL
	if err := util.ConnectToMemSQL(config); err != nil {
		fmt.Printf("Connection to MemSQL Failed. Please check the configuration and try again.\n")
		log.Fatalf("Failed to connect to MemSQL: %s", err)
	}
	fmt.Println("Connected to MemSQL")

	// Run health checks
	fmt.Println("Running Health Check")
	if err := steps.PreUpgrade(); err != nil {
		fmt.Printf("Health Check Failed. Please check the logs for more information.\n")
		log.Fatalf("Health Check Failed: %s\n", err)
	}
	fmt.Println("Health Check Complete")

	// Snapshot user databases
	fmt.Println("Taking Snapshots")
	if err := steps.SnapshotDatabases(); err != nil {
		fmt.Printf("Snapshot Failed. Please check the logs for more information.\n")
		log.Fatalf("Snapshot Failed: %s\n", err)
	}
	fmt.Println("All Snapshots Complete")

	// Update config prior to upgrade
	fmt.Println("Updating Configs")
	if err := steps.UpdateConfig("OFF"); err != nil {
		log.Fatalf("Update Failed: %s\n", err)
	}
	fmt.Println("All Configs updated")

}
