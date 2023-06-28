package rpc

import (
	"context"
)

type Host struct {
	HostID            string `json:"hostid"`
	Name              string `json:"name,omitempty"`
	MaintenanceFrom   string `json:"maintenance_from,omitempty"`
	MaintenanceStatus string `json:"maintenance_status,omitempty"`
	MaintenanceType   string `json:"maintenance_type,omitempty"`
	MaintenanceID     string `json:"maintenanceid,omitempty"`
}

var selectHosts = []string{"hostid", "name", "maintenance_from",
	"maintenance_status", "maintenance_type", "maintenanceid"}

func (c *Client) GetHostsByNamesFullMatch(ctx context.Context,
	names []string) ([]Host, error) {
	type Names struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: selectHosts,
		Filter: Names{
			Name: names,
		},
	}
	var hosts []Host
	if err := c.Client.Call(ctx, "host.get", params, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (c *Client) GetHostsByHostIDs(ctx context.Context,
	hostIDs []string) ([]Host, error) {
	params := struct {
		Output  any `json:"output"`
		HostIDs any `json:"hostids"`
	}{
		Output:  selectHosts,
		HostIDs: hostIDs,
	}
	var hosts []Host
	if err := c.Client.Call(ctx, "host.get", params, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (c *Client) GetHostsByGroupIDs(ctx context.Context,
	groupIDs []string) ([]Host, error) {
	params := struct {
		Output   any `json:"output"`
		GroupIDs any `json:"groupids"`
	}{
		Output:   selectHosts,
		GroupIDs: groupIDs,
	}
	var hosts []Host
	if err := c.Client.Call(ctx, "host.get", params, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}
