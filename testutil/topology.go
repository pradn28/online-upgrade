package testutil

import (
	"os"
	"testing"

	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/require"
)

var (
	MEMSQL_VERSION_1 = os.Getenv("MEMSQL_VERSION_5_5_12")
	MEMSQL_VERSION_2 = os.Getenv("MEMSQL_VERSION_5_7_2")
)

func init() {
	if MEMSQL_VERSION_1 == "" || MEMSQL_VERSION_2 == "" {
		panic("Must define env vars MEMSQL_VERSION_5_5_12 and MEMSQL_VERSION_5_7_2")
	}
}

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
		"--version-hash", MEMSQL_VERSION_1,
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
		"--version-hash", MEMSQL_VERSION_1,
		"--async",
	)

	// Deploy Leaf
	MustRun(t,
		"memsql-ops",
		"memsql-deploy",
		"--role", "leaf",
		"--port", "3307",
		"--version-hash", MEMSQL_VERSION_1,
	)

	require.Nil(t, util.OpsWaitMemsqlsOnlineConnected(2))

	return func() { TerminateCluster(t) }
}
