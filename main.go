package main

import (
	"fmt"
	"log"
	"os"

	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/util"
)

func main() {
	// Grab any config information specified on the command line
	// If no config information is passed in, we will use the defaults in 'config.go'
	config := util.ParseFlags()

	// create a channel for signals
	sigChan := make(chan os.Signal, 1)
	util.CatchSignals(sigChan, os.Interrupt)

	// Open log file
	logFinalizer, err := util.SetupLogging(config)
	if err != nil {
		panic(fmt.Errorf("Failed to setup logging: %s", err))
	}

	util.GetUserConfirmation("memsql-online-upgrade will execute various steps to upgrade your cluster. While there are steps to check your cluster is healthy, it is recommended that you backup and double check your cluster is in a healthy status prior to starting the upgrade.", "Type START to begin the upgrade: ", "START")

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
	fmt.Println("All Health Checks Completed")

	util.GetUserConfirmation("Your cluster appears to be healthy. In the next step we will snapshot all user databases and update configs.", "Type GO to continue: ", "GO")

	// Snapshot user databases
	fmt.Println("Taking Snapshots")
	if err := steps.SnapshotDatabases(); err != nil {
		fmt.Printf("Snapshot Failed. Please check the logs for more information.\n")
		log.Fatalf("Snapshot Failed: %s\n", err)
	}
	fmt.Println("All Snapshots Completed")

	// Update config prior to upgrade
	fmt.Println("Updating Configs")
	if err := steps.UpdateConfig("OFF"); err != nil {
		log.Fatalf("Update Failed: %s\n", err)
	}
	fmt.Println("All Configs updated")

	util.GetUserConfirmation("The next step is to upgrade the leaves. We will start with the first availability group and wait for confirmation before continuing to the next ", "Type UPGRADE to continue: ", "UPGRADE")

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

		if group == 1 {
			util.GetUserConfirmation("All leaves in Availability Group 1 are upgraded. Ready to upgrade leaves in Availability Group 2?", "Type NEXT to continue: ", "NEXT")
		}
	}

	// Rebalance cluster
	util.GetUserConfirmation("Next we will rebalance the cluster", "Type REBALANCE to continue: ", "REBALANCE")

	fmt.Println("Rebalancing user databases")
	if err := steps.RebalancePartitions(); err != nil {
		fmt.Println("Failed while rebalancing partitions. Please check the logs for more information.\n", err)
		log.Fatalf("Rebalance Partitions Failed: %s\n", err)
	}
	fmt.Println("Rebalance partitions completed successfully")

	// Upgrade aggregators
	fmt.Println("\nUpgrading Child Aggregators")
	if err := steps.UpgradeAggregators(); err != nil {
		fmt.Println("Failed while upgrading Aggregators. Please check the logs for more information.\n", err)
		log.Fatalf("Upgrade Failed: %s\n", err)
	}
	fmt.Println("All Child Aggregators have upgraded successfully")

	// Upgrade Master
	fmt.Println("Upgrading Master Aggregator")
	util.GetUserConfirmation("We are ready to upgrade the Master Aggregator. After the Master Aggregator is upgraded, we will re-set the configuration variables modified at the start of the upgrade.", "Type MASTER to continue: ", "MASTER")
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
