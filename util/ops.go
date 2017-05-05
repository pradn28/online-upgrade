package util

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
)

// Set Custom error messages
var (
	ErrNoMaster = errors.New("util/ops: Master Aggregator not found")
)

// AgentInfo struct for memsql-ops agent-list
type AgentInfo struct {
	AgentID string `json:"agent_id"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Role    string `json:"role"`
	State   string `json:"state"`
	Version string `json:"version"`
}

// MemsqlInfo struct for memsql-ops memsql-list
type MemsqlInfo struct {
	MemsqlID string `json:"memsql_id"`

	AgentID       string `json:"agent_id"`
	MemsqlVersion string `json:"memsql_version"`
	Role          string `json:"role"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Group         int    `json:"availability_group"`
	ClusterState  string `json:"cluster_state"`
	RunState      string `json:"run_state"`
	State         string `json:"state"`
}

// OpsAgentList returns a slice of agent-list
func OpsAgentList() ([]AgentInfo, error) {
	var infos []AgentInfo

	err := sh.Command(
		"memsql-ops",
		"agent-list",
		"--json",
	).UnmarshalJSON(&infos)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

// OpsMemsqlList return a slice of memsql-list
// By default we expect json. Do not pass in `-q`
func OpsMemsqlList(args ...string) ([]MemsqlInfo, error) {
	var infos []MemsqlInfo

	// Default Args
	listArgs := []string{"memsql-list", "--json"}
	// Append optional args
	listArgs = append(listArgs, args...)

	err := sh.Command("memsql-ops", listArgs).UnmarshalJSON(&infos)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// GetNodeIDs returns a slice of MemSQL node IDs by type (e.g. MASTER)
func GetNodeIDs(nodeType string) ([]string, error) {
	var memsqlIDs []string

	// Get a list of MemSQL nodes
	nodes, err := OpsMemsqlList("--memsql-role", nodeType)
	if err != nil {
		return nil, err
	}

	// Loop through nodes and return a slice of memsqlIDs
	for i := range nodes {
		node := nodes[i]
		memsqlIDs = append(memsqlIDs, node.MemsqlID)
	}
	return memsqlIDs, nil
}

// OpsMemsqlUpdateConfig updates memsql.cnf
// Accepts memsqlID, key, value, and a list of options
// We return output of the command and nill if not err
func OpsMemsqlUpdateConfig(memsqlID, key, value string, option ...string) error {

	args := []string{"memsql-update-config", memsqlID, "--key", key, "--value", value}
	args = append(args, option...)

	log.Printf("Running: %s", args)
	out, err := sh.Command("memsql-ops", args).Output()
	log.Print(string(out))

	if err != nil {
		return err
	}
	return nil
}

// OpsNodeManagement executes memsql-ops commands which
// take a single memsql_id as a positional argument.
// (start, stop ...)
func OpsNodeManagement(command, memsqlID string, arguments ...string) error {
	c := []string{command, memsqlID}
	c = append(c, arguments...)

	log.Printf("Running memsql-ops %s", c)
	out, err := sh.Command("memsql-ops", c).Output()
	log.Print(string(out))
	if err != nil {
		return err
	}
	return nil
}

// OpsMemsqlUpgrade upgrades a specified memsql node
func OpsMemsqlUpgrade(memsqlID string, arguments ...string) error {
	// Grab config data
	config := ParseFlags()

	u := []string{"memsql-upgrade", "--memsql-id", memsqlID}
	u = append(u, arguments...)

	if len(config.VersionHash) > 0 {
		u = append(u, "--version-hash", config.VersionHash)
	}
	if config.SkipVersionCheck == true {
		u = append(u, "--skip-version-check")
	}

	log.Printf("Running upgrade with %v", u)

	out, err := sh.Command("memsql-ops", u).Output()
	log.Print(string(out))
	if err != nil {
		return err
	}
	return nil
}

// OpsMemsqlGetVersion returns MemSQL Version for provided memsql_id
func OpsMemsqlGetVersion(memsqlID string) (string, error) {

	memsqls, err := OpsMemsqlList()
	if err != nil {
		return "", err
	}
	for i := range memsqls {
		v := memsqls[i]
		if memsqlID == v.MemsqlID {
			version := v.MemsqlVersion
			return version, err
		}
	}
	return "", err
}

// OpsWaitMemsqlsOnlineConnected checks and waits for the state
// of a specified number of nodes to be online and connected
func OpsWaitMemsqlsOnlineConnected(numNodes int) error {
	return StateChange{
		Target:  true,
		Timeout: time.Second * 120,
		Refresh: func() (interface{}, error) {
			infos, err := OpsMemsqlList()
			if err != nil {
				return false, err
			}
			if len(infos) < numNodes {
				return false, nil
			}
			for i := range infos {
				info := infos[i]
				if info.State != "ONLINE" {
					return false, nil
				}
				if info.ClusterState != "CONNECTED" {
					return false, nil
				}
			}
			return true, nil
		},
	}.WaitForState()
}

// OpsVersionCheck return Memsql Ops Version
// memsql-ops version --client-version-only
func OpsVersionCheck() ([]string, error) {
	out, err := sh.Command(
		"memsql-ops",
		"version",
		"--client-version-only",
	).Output()

	// Split on dot
	version := strings.Split(string(out), ".")

	if err != nil {
		return nil, err
	}
	return version, nil
}
