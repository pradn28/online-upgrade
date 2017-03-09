package steps

import (
	"errors"
	"fmt"
	"log"

	"github.com/memsql/online-upgrade/util"
)

// Set Custom error messages
var (
	ErrRedundancyLevel = errors.New("preUpgrade: Cluster must be configured with redundancy_level=2")
)

// PreUpgrade ensures that the MemSQL cluster is healthy
func PreUpgrade() error {
	log.Printf("PreUpgrade Check Started")

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

	log.Printf("PreUpgrade Completed Successfully")

	return nil
}
