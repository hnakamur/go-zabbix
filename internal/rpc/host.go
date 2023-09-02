package rpc

import (
	"context"
	"fmt"
	"slices"
	"strings"
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

	var notFoundNames []string
	for _, name := range names {
		if !slices.ContainsFunc(hosts, func(host Host) bool {
			return host.Name == name
		}) {
			notFoundNames = append(notFoundNames, name)
		}
	}
	if len(notFoundNames) > 0 {
		return nil, fmt.Errorf("hosts not found: %s", strings.Join(notFoundNames, ", "))
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

	var notFoundHostIDs []string
	for _, hostID := range hostIDs {
		if !slices.ContainsFunc(hosts, func(host Host) bool {
			return host.HostID == hostID
		}) {
			notFoundHostIDs = append(notFoundHostIDs, hostID)
		}
	}
	if len(notFoundHostIDs) > 0 {
		return nil, fmt.Errorf("host IDs not found: %s", strings.Join(notFoundHostIDs, ", "))
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
