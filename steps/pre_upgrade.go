package steps

import (
	"fmt"
	"github.com/memsql/online-upgrade/util"
)

func stepPreUpgrade() error {
	// is HA setup
	redundancy_level, err := util.DBGetVariable("redundancy_level")
	if err != nil {
		return err
	}

	if redundancy_level != "2" {
		return fmt.Errorf("Cluster must be configured with redundancy_level=2")
	}

	// ops is happy + healthy
	agents, err := util.OpsAgentList()
	if err != nil {
		return err
	}

	for i := range agents {
		agent := agents[i]
		if agent.State != "ONLINE" {
			return fmt.Errorf("Agent %s is offline", agent.AgentID)
		}
	}

	// all nodes are online
	memsqls, err := util.OpsMemsqlList()
	if err != nil {
		return err
	}
	if len(memsqls) == 0 {
		return fmt.Errorf("No MemSQL nodes found")
	}

	for i := range memsqls {
		memsql := memsqls[i]
		if memsql.State != "ONLINE" {
			return fmt.Errorf("MemSQL Node %s is offline", memsql.MemsqlID)
		}
		if memsql.ClusterState != "CONNECTED" {
			return fmt.Errorf("MemSQL Node %s is not connected to the cluster", memsql.MemsqlID)
		}
	}

	return nil
}
