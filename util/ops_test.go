package util_test

import (
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOps(t *testing.T) {
	defer testutil.ClusterInABox(t)()

	t.Run("OpsAgentList", func(t *testing.T) {
		agents, err := util.OpsAgentList()
		assert.Nil(t, err)
		assert.Len(t, agents, 1)
		assert.Equal(t, "ONLINE", agents[0].State)
	})

	t.Run("OpsMemsqlList", func(t *testing.T) {
		memsqls, err := util.OpsMemsqlList()
		assert.Nil(t, err)
		assert.Len(t, memsqls, 2)
		assert.Equal(t, "ONLINE", memsqls[0].State)
		assert.Equal(t, "CONNECTED", memsqls[0].ClusterState)
		assert.Equal(t, "MASTER", memsqls[0].Role)
		assert.Equal(t, "ONLINE", memsqls[1].State)
		assert.Equal(t, "CONNECTED", memsqls[1].ClusterState)
		assert.Equal(t, "LEAF", memsqls[1].Role)
	})

	t.Run("OpsWaitMemsqlsOnlineConnected", func(t *testing.T) {
		testutil.MustRun(t, "memsql-ops", "memsql-restart", "--all", "--async")
		assert.Nil(t, util.OpsWaitMemsqlsOnlineConnected(2))
	})
}