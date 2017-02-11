package steps

import (
	"fmt"
	"github.com/memsql/online-upgrade/util"
)

func stepSnapshotDatabases() error {
	dbs, err := util.DBGetUserDatabases()
	if err != nil {
		return err
	}
	for i := range dbs {
		db := dbs[i]
		err := util.DBSnapshotDatabase(db)
		if err != nil {
			return fmt.Errorf("Failed to snapshot database %s: %s", db, err)
		}
	}
	return nil
}
