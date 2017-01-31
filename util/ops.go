package util

import (
	"github.com/codeskyblue/go-sh"
	"time"
)

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

func OpsWaitMemsqlOnline() error {
	return StateChange{
		Target:  true,
		Timeout: time.Second * 60,
		Refresh: func() (interface{}, error) {
			infos, err := OpsMemsqlList()
			if err != nil {
				return false, err
			}
			for i := range infos {
				info := infos[i]
				if info.State != "online" {
					return false, nil
				}
			}
			return true, nil
		},
	}.WaitForState()
}
