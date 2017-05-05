package steps

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/memsql/online-upgrade/util"
)

type LeafPair struct {
	PairHost string
	PairPort int
}

// Define custom errors
var (
	errMasterNotFound = errors.New("detach_leaf: Missing Master after detaching partition")
)

// DetachLeaves will detach leaves from a specified AG group
func DetachLeaves(group int) error {
	// Create new spinner to show activity of detach
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	// Get partitions from Show Cluster Status
	partitions, err := util.DBShowClusterStatus()
	if err != nil {
		return err
	}
	log.Printf("Detaching leaves in Availability Group %d", group)

	// Check for Oprhan Partitions
	if OrphanCheckErr := util.OrphanCheck(partitions); err != nil {
		return OrphanCheckErr
	}
	// Check that each partition has a Master and Slave
	if MasterSlaveCheckErr := util.MasterSlaveCheck(partitions); err != nil {
		return MasterSlaveCheckErr
	}

	// Get a list of leaves
	leaves, err := util.DBShowLeaves()
	if err != nil {
		return err
	}

	// Gather a slice on leaves that were detach to cross reference after detached
	detachedPairs := []*LeafPair{}

	// Detach leaves in a specified AG (host, port)
	for i := range leaves {
		l := leaves[i]
		// If leaf is in the specified AG, detach it
		if l.AG == group {

			// Configure spinner
			s.Prefix = " "
			s.Suffix = fmt.Sprintf(" Detaching %s:%d", l.Host, l.Port)
			s.FinalMSG = fmt.Sprintf(" ✓ Detach complete for %s:%d\n", l.Host, l.Port)
			s.Start()

			// Detach leaf
			detachErr := util.DBDetachLeaf(l.Host, l.Port)
			if err != nil {
				return detachErr
			}
			// Add host pair to detachedPairs slice
			pair := LeafPair{l.PairHost, l.PairPort}
			detachedPairs = append(detachedPairs, &pair)
			log.Printf("Detached %s:%d", l.PairHost, l.PairPort)
			s.Stop()
		}
	}

	// Verify all promotions completed successfully
	// Configure spinner
	s.Prefix = " "
	s.Suffix = fmt.Sprint(" Verifing promotions")
	s.FinalMSG = fmt.Sprintln(" ✓ All partitions promoted")
	s.Start()
	time.Sleep(5 * time.Second)

	// Get updated Show Cluster Status
	clusterStatus, err := util.DBShowClusterStatus()
	if err != nil {
		return err
	}

	// Loop through all leaves that have been detached
	for i := range detachedPairs {
		d := detachedPairs[i]
		// for each leaf that was detached verify that it is now the master
		for i := range clusterStatus {
			p := clusterStatus[i]
			if d.PairHost == p.Host && d.PairPort == p.Port && p.Role != "Master" && p.Role != "Reference" {
				log.Printf("Partition not promoted:: %s:%d [%s] ", p.Host, p.Port, p.Role)
				return errMasterNotFound
			}
		}
	}
	log.Printf("Availability Group %d detached successfully", group)
	s.Stop()

	return nil
}
