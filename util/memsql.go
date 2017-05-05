package util

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	errMasterSlaveMismatch = errors.New("detach_leaf: Missing partition")
	errOrphanFound         = errors.New("detach_leaf: Orphan partition found")
)

// Leaf - struct for leaf data return from query
type Leaf struct {
	Host     string  `db:"Host"`
	Port     int     `db:"Port"`
	PairHost string  `db:"Pair_Host"`
	PairPort int     `db:"Pair_Port"`
	AG       int     `db:"Availability_Group"`
	OpenConn int     `db:"Opened_Connections"`
	AvgRT    []uint8 `db:"Average_Roundtrip_Latency_ms"`
	State    string  `db:"State"`
}

// ClusterStatus - struct for data returned from query
type ClusterStatus struct {
	Host     string  `db:"Host"`
	Port     int     `db:"Port"`
	Database string  `db:"Database"`
	Role     string  `db:"Role"`
	State    string  `db:"State"`
	Position []uint8 `db:"Position"`
	Details  string  `db:"Details"`
}

// RestoreRedundancy data return from Exlpain Restore Redundancy
type RestoreRedundancy struct {
	Action     string `db:"Action"`
	Ordinal    int    `db:"Ordinal"`
	TargetHost string `db:"Target_Host"`
	TargetPort int    `db:"Target_Port"`
	Phase      int    `db:"Phase"`
}

var dbConn *sqlx.DB

// ConnectToMemSQL does just that
func ConnectToMemSQL(config Config) error {
	var err error

	connParams := strings.Join([]string{
		// convert timestame and date to time.Time
		"parseTime=true",
		// don't use the binary protocol
		"interpolateParams=true",
		// set a sane connection timeout rather than the default infinity
		"timeout=10s",
	}, "&")

	connString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/information_schema?%s",
		config.MasterUser,
		config.MasterPass,
		config.MasterHost,
		config.MasterPort,
		connParams,
	)

	log.Printf("Connecting to MemSQL %s", connString)

	dbConn, err = sqlx.Open("mysql", connString)
	if err != nil {
		return err
	}

	dbConn.SetConnMaxLifetime(time.Hour * 6)
	dbConn.SetMaxIdleConns(10)

	return dbConn.Ping()
}

// TestGetDB exposes the private dbConn variable for testing.
// Should not be used in any of the actual code - all queries should be
// materialized as tested functions in this package instead.
func TestGetDB(t *testing.T) *sqlx.DB {
	if dbConn == nil {
		t.Fatal("ConnectToMemSQL() must run before TestGetDB() can be called")
	}

	return dbConn
}

// DBGetVariable returns value of variable name specified
func DBGetVariable(varName string) (string, error) {
	var res struct {
		Name  string `db:"Variable_name"`
		Value string `db:"Value"`
	}
	err := dbConn.Get(&res, "SHOW VARIABLES LIKE ?", varName)
	return res.Value, err
}

// DBRowCount returns a count for specified table
func DBRowCount(database, table string) (string, error) {
	var res struct {
		Count string `db:"count"`
	}
	err := dbConn.Get(&res, fmt.Sprintf("select count(*) as count from `%s`.`%s`", database, table))
	if err != nil {
		return "", err
	}
	return res.Count, err
}

// DBSetVariable sets database variables
func DBSetVariable(db string) error {
	_, err := dbConn.Exec(db)
	return err
}

// DBGetUserDatabases returns all user databases
func DBGetUserDatabases() ([]string, error) {
	rows, err := dbConn.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	dbs := make([]string, 0)
	for rows.Next() {
		var db string
		err := rows.Scan(&db)
		if err != nil {
			return nil, err
		}
		// Check for DR sharding db
		drSharding, _ := regexp.MatchString("^sharding_.+$", db)

		if db != "information_schema" &&
			db != "memsql" &&
			db != "sharding" &&
			drSharding == false {
			dbs = append(dbs, db)
		}

	}
	return dbs, nil
}

// DBSnapshotDatabase executes a snapshot of provided databse
func DBSnapshotDatabase(db string) error {
	_, err := dbConn.Exec(fmt.Sprintf("SNAPSHOT DATABASE `%s`", db))
	return err
}

// DBShowLeaves returns the output of show leaves
// Return a slice of pointers to the Leaf struct
func DBShowLeaves() ([]*Leaf, error) {

	leaves := []*Leaf{}
	rows, err := dbConn.Queryx("SHOW LEAVES")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		leaf := new(Leaf)
		err := rows.StructScan(&leaf)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}

	return leaves, nil
}

// DBShowClusterStatus returns output of show cluster status
// Return a slice of pointers to the ClusterStatus struct
func DBShowClusterStatus() ([]*ClusterStatus, error) {

	partitions := []*ClusterStatus{}
	rows, err := dbConn.Queryx("SHOW CLUSTER STATUS")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		p := new(ClusterStatus)
		err := rows.StructScan(&p)
		if err != nil {
			return nil, err
		}
		partitions = append(partitions, p)
	}

	return partitions, nil
}

// DBDetachLeaf will detach a specified host (DETACH LEAF 'host':port)
func DBDetachLeaf(host string, port int) error {
	_, err := dbConn.Exec(fmt.Sprintf("DETACH LEAF `%s`:%d", host, port))
	if err != nil {
		return err
	}
	return err

}

// DBAttachLeaf will attach a specified host (ATTACH LEAF 'host':port)
func DBAttachLeaf(host string, port int) error {
	_, err := dbConn.Exec(fmt.Sprintf("ATTACH LEAF `%s`:%d NO REBALANCE", host, port))
	if err != nil {
		return err
	}
	return err

}

// DBRestoreRedundancy will restore redundancy on specified DB
func DBRestoreRedundancy(database string) error {
	_, err := dbConn.Exec(fmt.Sprintf("RESTORE REDUNDANCY on %s", database))
	if err != nil {
		return err
	}
	return err
}

// DBExplainRestoreRedundancy will run explain restore redundancy on specified DB
func DBExplainRestoreRedundancy(database string) ([]*RestoreRedundancy, error) {

	actions := []*RestoreRedundancy{}
	rows, err := dbConn.Queryx(fmt.Sprintf("EXPLAIN RESTORE REDUNDANCY on %s", database))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		a := new(RestoreRedundancy)
		err := rows.StructScan(&a)
		if err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, err
}

// DBRebalancePartitions will run rebalance on specified DB
func DBRebalancePartitions(database string) error {
	_, err := dbConn.Exec(fmt.Sprintf("REBALANCE PARTITIONS on %s", database))
	if err != nil {
		return err
	}
	return err
}

// MasterSlaveCheck will check each partition has a Master and Slave
// masterSlaveCheck Requires a slice if partitions from Cluster Status
// util.ClusterStatus will return a slice of pointers to the struct
func MasterSlaveCheck(partitions []*ClusterStatus) error {
	for a := range partitions {
		p1 := partitions[a]
		roles := [][]string{[]string{"Master", "Slave"}, []string{"Slave", "Master"}}

		for i := range roles {
			x := roles[i][0]
			y := roles[i][1]
			if p1.Role == x {
				// Find the Slave
				var paired bool
				for b := range partitions {
					p2 := partitions[b]
					if p2.Database == p1.Database && p2.Role == y {
						paired = true
					}
				}
				if paired != true {
					log.Printf("%s Partition not found for %s", y, p1.Database)
					return errMasterSlaveMismatch
				}
			}
		}
	}
	return nil
}

// OrphanCheck will check for any orphan partitions
// orphanCheck requires a slice if partitions from Cluster Status
// util.ClusterStatus will return a slice of pointers to the struct
func OrphanCheck(partitions []*ClusterStatus) error {
	for a := range partitions {
		p1 := partitions[a]
		// Check for Oprhan Partitions
		if p1.Role == "Orphan" {
			log.Printf("Orphan partition - %s [%s]\n",
				p1.Database, p1.Role,
			)
			return errOrphanFound
		}
	}
	return nil
}
