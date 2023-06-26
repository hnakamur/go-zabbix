package main

import (
	"context"

	"github.com/hnakamur/go-zabbix"
)

type Host struct {
	HostID zabbix.ID `json:"hostid"`
	Name   string    `json:"name,omitempty"`
}

func (c *myClient) GetHostsByNamesFullMatch(ctx context.Context,
	names []string) ([]Host, error) {
	type Names struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: []string{"hostid", "name"},
		Filter: Names{
			Name: names,
		},
	}
	var result []Host
	if err := c.Client.Call(ctx, "host.get", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
