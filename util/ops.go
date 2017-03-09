package util

import (
	"errors"
	"log"
	"time"

	"github.com/codeskyblue/go-sh"
)

// Set Custom error messages
var (
	ErrNoMaster = errors.New("util/ops: Master Aggregator not found")
)

type agentInfo struct {
	AgentID string `json:"agent_id"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Role    string `json:"role"`
	State   string `json:"state"`
	Version string `json:"version"`
}

type memsqlInfo struct {
	MemsqlID string `json:"memsql_id"`

	AgentID       string `json:"agent_id"`
	MemsqlVersion string `json:"memsql_version"`
	Role          string `json:"role"`
	Host          string `json:"host"`
	Port          int    `json:"port"`

	ClusterState string `json:"cluster_state"`
	RunState     string `json:"run_state"`
	State        string `json:"state"`
}

func OpsAgentList() ([]agentInfo, error) {
	var infos []agentInfo

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

func OpsMemsqlList() ([]memsqlInfo, error) {
	var infos []memsqlInfo

	err := sh.Command(
		"memsql-ops",
		"memsql-list",
		"--json",
	).UnmarshalJSON(&infos)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

// GetNodeIDs returns a slice of MemSQL node IDs by type (e.g. MASTER)
func GetNodeIDs(nodeType string) ([]string, error) {
	var memsqlIDs []string

	// Get a list of MemSQL nodes
	nodes, err := OpsMemsqlList()
	if err != nil {
		return nil, err
	}

	// Loop through nodes and return memsqlIDs
	for i := range nodes {
		node := nodes[i]
		if node.Role == nodeType {
			memsqlIDs = append(memsqlIDs, node.MemsqlID)
		}
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

func OpsWaitMemsqlsOnlineConnected(numNodes int) error {
	return StateChange{
		Target:  true,
		Timeout: time.Second * 60,
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
