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

	// Upgrade Leaves
	availabilityGroups := []int{1, 2}

	for i := range availabilityGroups {
		group := availabilityGroups[i]

		// Detach leaves
		fmt.Printf("Detaching Leaves in Availability Group %d\n", group)
		if err := steps.DetachLeaves(group); err != nil {
			fmt.Println("Failed while detaching leaves. Please check the logs for more information.\n", err)
			log.Fatalf("Detaching Failed: %s\n", err)
		}
		fmt.Printf("All leaves in Availability Group %d detached successfully\n", group)

		// Upgrade leaves
		fmt.Printf("Upgrading Leaves in Availability Group %d\n", group)
		if err := steps.UpgradeLeaves(group); err != nil {
			fmt.Println("Failed while upgrading leaves. Please check the logs for more information.\n", err)
			log.Fatalf("Upgrade Failed: %s\n", err)
		}
		fmt.Printf("All leaves in Availability Group %d have upgraded successfully\n", group)

		// Attach leaves
		fmt.Printf("Attaching Leaves in Availability Group %d\n", group)
		if err := steps.AttachLeaves(group); err != nil {
			fmt.Println("Failed while attaching leaves. Please check the logs for more information.\n", err)
			log.Fatalf("Attaching Failed: %s\n", err)
		}
		fmt.Printf("All leaves in Availability Group %d attached successfully\n", group)

		// Restore redundancy
		fmt.Println("Restoring redundancy on user databases")
		if err := steps.RestoreRedundancy(); err != nil {
			fmt.Println("Failed while restoring redundancy. Please check the logs for more information.\n", err)
			log.Fatalf("Restore Redundancy: %s\n", err)
		}
		fmt.Println("Redundancy restored successfully")
	}

	// Rebalance cluster
	fmt.Println("Rebalancing user databases")
	if err := steps.RebalancePartitions(); err != nil {
		fmt.Println("Failed while rebalancing partitions. Please check the logs for more information.\n", err)
		log.Fatalf("Rebalance Partitions Failed: %s\n", err)
	}
	fmt.Println("Rebalance partitions completed successfully")

	// Upgrade aggregators
	fmt.Println("Upgrading Child Aggregators")
	if err := steps.UpgradeAggregators(); err != nil {
		fmt.Println("Failed while upgrading Aggregators. Please check the logs for more information.\n", err)
		log.Fatalf("Upgrade Failed: %s\n", err)
	}
	fmt.Println("All Child Aggregators have upgraded successfully")

	// Upgrade Master
	fmt.Println("Upgrading Master Aggregator")
	if err := steps.UpgradeMaster(); err != nil {
		fmt.Println("Failed while upgrading Master. Please check the logs for more information.\n", err)
		log.Fatalf("Upgrade Failed: %s\n", err)
	}
	fmt.Println("Master Aggregator has been upgraded successfully")

	// Update config post upgrade
	fmt.Println("Updating Configs")
	if err := steps.UpdateConfig("ON"); err != nil {
		log.Fatalf("Update Failed: %s\n", err)
	}
	fmt.Println("All Configs updated")

	fmt.Println("Upgrade completed successfully")
}
