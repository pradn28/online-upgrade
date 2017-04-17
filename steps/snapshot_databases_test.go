package steps_test

import (
	"github.com/memsql/online-upgrade/steps"
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestSnapshotDatabases(t *testing.T) {
	//Build test cluster
	defer testutil.ClusterSinglebox(t)()

	var database = "testing"

	// Create a testing database
	require.Nil(t, util.ConnectToMemSQL(util.ParseFlags()))
	defer testutil.CreateDatabase(t, database)()

	t.Run("SnapshotDatabases", func(t *testing.T) {
		err := steps.SnapshotDatabases()
		assert.Nil(t, err)
	})

}
