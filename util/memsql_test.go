package util_test

import (
	"github.com/memsql/online-upgrade/util"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemSQL(t *testing.T) {
	util.ConnectToMemSQL(util.ParseFlags())
	res, err := util.DBGetVariable("version_compile_os")
	assert.Nil(t, err)
	assert.Equal(t, res, "Linux")
}
