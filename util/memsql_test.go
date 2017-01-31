package util_test

import (
	"github.com/memsql/online-upgrade/testutil"
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemsql(t *testing.T) {
	defer testutil.ClusterInABox(t)()

	util.ConnectToMemSQL(util.ParseFlags())
	res, err := util.DBGetVariable("version_compile_os")
	assert.Nil(t, err)
	assert.Equal(t, res, "Linux")
}
