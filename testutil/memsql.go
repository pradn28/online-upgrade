package testutil

import (
	"github.com/codeskyblue/go-sh"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/memsql/online-upgrade/util"
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

// ClusterInABox spins up a cluster-in-a-box cluster on this machine.
// Returns a finalizer to stop && delete the cluster.
func ClusterInABox(t *testing.T) func() {
	// Start Ops
	err := sh.Command("memsql-ops", "start").Run()
	require.Nil(t, err)

	// Deploy Master
	err = sh.Command(
		"memsql-ops",
		"memsql-deploy",
		"--role", "master",
		"--version-hash", MEMSQL_VERSION_1,
	).Run()
	require.Nil(t, err)

	// Deploy Leaf
	err = sh.Command(
		"memsql-ops",
		"memsql-deploy",
		"--role", "leaf",
		"--port", "3307",
		"--version-hash", MEMSQL_VERSION_1,
	).Run()
	require.Nil(t, err)

	err = util.OpsWaitMemsqlOnlineConnected()
	require.Nil(t, err)

	return func() {
		// Stop nodes
		err := sh.Command(
			"memsql-ops",
			"memsql-stop",
			"--hard-stop",
			"--all",
		).Run()
		require.Nil(t, err)

		// Delete nodes
		err = sh.Command(
			"memsql-ops",
			"memsql-delete",
			"--delete-without-prompting",
			"--all",
		).Run()
		require.Nil(t, err)
	}
}
