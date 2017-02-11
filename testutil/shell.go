package testutil

import (
	"github.com/codeskyblue/go-sh"
	"github.com/stretchr/testify/require"
	"testing"
)

func MustRun(t *testing.T, name string, a ...interface{}) {
	require.Nil(t, sh.Command(name, a...).Run())
}
