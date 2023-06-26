package main

import (
	"context"

	"github.com/hnakamur/go-zabbix"
)

type HostGroup struct {
	GroupID zabbix.ID `json:"groupid"`
	Name    string    `json:"name,omitempty"`
}

func (c *myClient) GetHostGroupsByNamesFullMatch(ctx context.Context,
	names []string) ([]HostGroup, error) {
	type Names struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: []string{"groupid", "name"},
		Filter: Names{
			Name: names,
		},
	}
	var result []HostGroup
	if err := c.Client.Call(ctx, "hostgroup.get", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
