package testutil

import (
	"os"
	"testing"

	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/require"
)

var (
	memsqlVersion1 = os.Getenv("MEMSQL_VERSION_5_5_12")
	memsqlVersion2 = os.Getenv("MEMSQL_VERSION_5_7_2")
)

func init() {
	if memsqlVersion1 == "" || memsqlVersion2 == "" {
		panic("Must define env vars MEMSQL_VERSION_5_5_12 and MEMSQL_VERSION_5_7_2")
	}
}

// TerminateCluster does exactly what you would think
func TerminateCluster(t *testing.T) {
	// Stop nodes
	MustRun(t,
		"memsql-ops",
		"memsql-stop",
		"--hard-stop",
		"--all",
	)

	// Delete nodes
	MustRun(t,
		"memsql-ops",
		"memsql-delete",
		"--delete-without-prompting",
		"--all",
	)
}

// ClusterSinglebox spins up a single leaf
// Returns a finalizer to stop && delete the node
func ClusterSinglebox(t *testing.T) func() {
	// Start Ops, if it's not already running
	MustRun(t, "memsql-ops", "start")

	// Deploy leaf
	MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--role", "leaf",
		"--version-hash", memsqlVersion1,
	)

	return func() { TerminateCluster(t) }
}

// ClusterInABox spins up a cluster-in-a-box cluster on this machine.
// Returns a finalizer to stop && delete the cluster.
func ClusterInABox(t *testing.T) func() {
	// Start Ops
	MustRun(t, "memsql-ops", "start")

	// Deploy Master async
	MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--role", "master",
		"--version-hash", memsqlVersion1,
		"--async",
	)

	// Deploy Leaf
	MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--role", "leaf",
		"--port", "3307",
		"--version-hash", memsqlVersion1,
	)

	require.Nil(t, util.OpsWaitMemsqlsOnlineConnected(2))

	return func() { TerminateCluster(t) }
}

// ClusterHA builds a high availablity cluster to test HA test
func ClusterHA(t *testing.T) func() {
	// Start Ops
	MustRun(t, "memsql-ops", "start")

	// Get agent count
	agentList, _ := util.OpsAgentList()
	agentCount := len(agentList)

	if agentCount == 1 {
		// Start remote agent
		MustRun(t,
			"sshpass", "-e",
			"ssh", "-oStrictHostKeyChecking=no",
			"root@online-upgrade-child",
			"memsql-ops", "start",
		)
		// Set remote agent to follow PRIMARY
		MustRun(t,
			"sshpass", "-e",
			"ssh", "-oStrictHostKeyChecking=no",
			"root@online-upgrade-child",
			"memsql-ops", "follow",
			"-h", "online-upgrade-master",
		)
	}

	// Get agent IDs
	agentInfo, _ := util.OpsAgentList()

	var masterID string
	var childID string

	for i := range agentInfo {
		a := agentInfo[i]
		if a.Role == "PRIMARY" {
			masterID = a.AgentID
		} else if a.Role == "FOLLOWER" {
			childID = a.AgentID
		}
	}

	// Deploy Master async
	MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--agent-id", masterID,
		"--role", "master",
		"--version-hash", memsqlVersion1,
		"--async",
	)
	// Deploy Child Aggregator async
	MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--role", "aggregator",
		"--agent-id", childID,
		"--version-hash", memsqlVersion1,
		"--async",
	)

	// Deploy first Leaf on Master
	go MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--agent-id", masterID,
		"--role", "leaf",
		"--port", "3307",
		"--version-hash", memsqlVersion1,
	)

	// Deploy third Leaf on Child
	go MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--role", "leaf",
		"--port", "3307",
		"--agent-id", childID,
		"--version-hash", memsqlVersion1,
	)

	// Wait for leaves to be online and connected before enabling HA
	require.Nil(t, util.OpsWaitMemsqlsOnlineConnected(4))

	// enable high availability
	MustRun(t,
		"memsql-ops",
		"memsql-enable-high-availability",
		"--no-prompt",
		"--skip-disk-check",
	)

	return func() { TerminateCluster(t) }
}
