package main

import (
	"context"
	"time"
)

type Host struct {
	HostID            string
	Name              string
	MaintenanceFrom   time.Time
	MaintenanceStatus string
	MaintenanceType   MaintenanceType
	MaintenanceID     string
}

type rpcHost struct {
	HostID            string `json:"hostid"`
	Name              string `json:"name,omitempty"`
	MaintenanceFrom   string `json:"maintenance_from,omitempty"`
	MaintenanceStatus string `json:"maintenance_status,omitempty"`
	MaintenanceType   string `json:"maintenance_type,omitempty"`
	MaintenanceID     string `json:"maintenanceid,omitempty"`
}

func toHost(h rpcHost) (Host, error) {
	maintenanceFrom, err := ParseTimestamp(h.MaintenanceFrom)
	if err != nil {
		return Host{}, err
	}

	return Host{
		HostID:            h.HostID,
		Name:              h.Name,
		MaintenanceFrom:   time.Time(maintenanceFrom),
		MaintenanceStatus: h.MaintenanceStatus,
		MaintenanceType:   MaintenanceType(h.MaintenanceType),
		MaintenanceID:     h.MaintenanceID,
	}, nil
}

func toRPCHost(h Host) (rpcHost, error) {
	return rpcHost{
		HostID: h.HostID,
		Name:   h.Name,
		// Keep empty values for readonly properties
	}, nil
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
		Output: selectHosts,
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
