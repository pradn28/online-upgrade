package steps

import (
	"github.com/memsql/online-upgrade/util"
)

// RestoreRedundancy to all user databases
func RestoreRedundancy() error {

	dbs, err := util.DBGetUserDatabases()
	if err != nil {
		return err
	}
	for i := range dbs {
		database := dbs[i]
		restoreErr := util.DBRestoreRedundancy(database)
		if restoreErr != nil {
			return restoreErr
		}
	}
	return err
}

// RebalancePartitions on all user databases
func RebalancePartitions() error {

	dbs, err := util.DBGetUserDatabases()
	if err != nil {
		return err
	}
	for i := range dbs {
		database := dbs[i]
		rebalanceErr := util.DBRebalancePartitions(database)
		if rebalanceErr != nil {
			return rebalanceErr
		}
	}
	return err
}
