package steps

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/memsql/online-upgrade/util"
)

// TODO: Check for unmonitored nodes
// TODO: Generate a cluster report. Add flag to disable `-no-report`
// TODO: Test for DR cluster.

// Set Custom error messages
var (
	ErrRedundancyLevel = errors.New("preUpgrade: Cluster must be configured with redundancy_level=2")
)

// PreUpgrade ensures that the MemSQL cluster is healthy
func PreUpgrade() error {
	// Create new spinner to show activity of health check
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = " "
	s.Suffix = fmt.Sprint(" Running Health Check")
	s.FinalMSG = fmt.Sprint(" ✓ Health Check complete\n")
	s.Start()
	defer s.Stop()

	log.Printf("PreUpgrade Check Started")

	// MemSQL Ops version Check
	version, _ := util.OpsVersionCheck()
	major, _ := strconv.Atoi(version[0])
	minor, _ := strconv.Atoi(version[1])

	// Ops 5.7 or higher is required
	if !(major >= 5 && minor >= 7) {
		return fmt.Errorf("Online upgrade requires MemSQL Ops Version 5.7 or higher. Found: %d.%d", major, minor)
	}
	fmt.Printf("MemSQL Ops Version = %d.%d", major, minor)

	// Check redundancy level. Redundancy level must be 2 (HA)
	redundancyLevel, err := util.DBGetVariable("redundancy_level")
	if err != nil {
		return err
	}
	if redundancyLevel != "2" {
		return ErrRedundancyLevel
	}
	log.Printf("Redundancy Level = %s", redundancyLevel)

	// Get agent list from MemSQL Ops
	agents, err := util.OpsAgentList()
	if err != nil {
		return err
	}
	// Verify all agents are online
	for i := range agents {
		agent := agents[i]
		if agent.State != "ONLINE" {
			return fmt.Errorf("preUpgrade: Agent %s is offline", agent.AgentID)
		}
		log.Printf("Agent %s is %s [%s:%d]", agent.AgentID, agent.State, agent.Host, agent.Port)
	}

	// Get a list of MemSQL nodes
	memsqls, err := util.OpsMemsqlList()
	if err != nil {
		return err
	}
	if len(memsqls) == 0 {
		return fmt.Errorf("preUpgrade: No MemSQL nodes found")
	}

	// Verify all nodes are online
	for i := range memsqls {
		memsql := memsqls[i]
		if memsql.State != "ONLINE" {
			return fmt.Errorf("preUpgrade: MemSQL Node %s is offline", memsql.MemsqlID)
		}
		if memsql.ClusterState != "CONNECTED" {
			return fmt.Errorf("preUpgrade: MemSQL Node %s is not connected to the cluster", memsql.MemsqlID)
		}
		log.Printf("MemSQL Node %s is %s and %s", memsql.MemsqlID, memsql.State, memsql.ClusterState)
	}

	// ExplainRestoreRedundancy for all user databases
	RestoreRedundancyCount := 0
	dbs, err := util.DBGetUserDatabases()
	if err != nil {
		return err
	}
	for i := range dbs {
		database := dbs[i]
		actions, restoreErr := util.DBExplainRestoreRedundancy(database)
		if restoreErr != nil {
			return restoreErr
		}
		RestoreRedundancyCount += len(actions)
	}
	if RestoreRedundancyCount > 0 {
		return fmt.Errorf("preUpgrade: Unbalanced Cluster. See EXPLAIN RESTORE REDUNDANCY")
	}

	log.Printf("PreUpgrade Completed Successfully")

	return nil
}
