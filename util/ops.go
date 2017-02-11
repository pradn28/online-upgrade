package util

import (
	"github.com/codeskyblue/go-sh"
	"time"
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
