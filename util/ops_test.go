package util_test

import (
	"os"

	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	versionHash = os.Getenv("MEMSQL_VERSION_5_7_2")
)

func TestOps(t *testing.T) {
	defer testutil.ClusterHA(t)()

	// Test we can receive a list of agents
	t.Run("OpsAgentList", func(t *testing.T) {
		agents, err := util.OpsAgentList()
		assert.Nil(t, err)
		assert.Len(t, agents, 2)
		assert.Equal(t, "ONLINE", agents[0].State)
	})

	// Test we are able to receive memsql-list
	t.Run("OpsMemsqlList", func(t *testing.T) {
		memsqls, err := util.OpsMemsqlList()
		assert.Nil(t, err)
		assert.Len(t, memsqls, 4)
		assert.Equal(t, "ONLINE", memsqls[0].State)
		assert.Equal(t, "MASTER", memsqls[0].Role)
		assert.Equal(t, "ONLINE", memsqls[1].State)
		assert.Equal(t, "AGGREGATOR", memsqls[1].Role)
		assert.Equal(t, "ONLINE", memsqls[2].State)
		assert.Equal(t, "LEAF", memsqls[2].Role)
	})

	// Test wait for online and connected nodes
	t.Run("OpsWaitMemsqlsOnlineConnected", func(t *testing.T) {
		testutil.MustRun(t, "memsql-ops", "memsql-restart", "--all", "--async")
		assert.Nil(t, util.OpsWaitMemsqlsOnlineConnected(4))
	})

	// Test we receive a node_id that matches Regexp
	t.Run("GetNodeIDs", func(t *testing.T) {
		nodeIDs, err := util.GetNodeIDs("LEAF")
		assert.Nil(t, err)
		assert.Regexp(t, "[0-9A-Z]{40}", nodeIDs[0])
	})

	// Test we can update configs on Master
	t.Run("OpsMemsqlUpdateConfig", func(t *testing.T) {
		memsqlID, _ := util.GetNodeIDs("MASTER")
		err := util.OpsMemsqlUpdateConfig(memsqlID[0], "auto_attach", "ON", "--set-global")
		assert.Nil(t, err)
	})

	// Test that we can stop a leaf
	t.Run("OpsNodeManagement", func(t *testing.T) {
		nodeIDs, _ := util.GetNodeIDs("LEAF")
		stopErr := util.OpsNodeManagement("memsql-stop", nodeIDs[0])
		assert.Nil(t, stopErr)
	})

	// Test upgrade
	t.Run("MemsqlUpgrade", func(t *testing.T) {
		nodeIDs, _ := util.GetNodeIDs("LEAF")
		upgradeErr := util.OpsMemsqlUpgrade(
			nodeIDs[0],
			"--no-prompt",
			"--skip-snapshot",
			"--no-backup-data-directories",
			"--version-hash", versionHash,
		)
		assert.Nil(t, upgradeErr)
	})

}
