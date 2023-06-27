package main

import (
	"context"
	"time"

	"github.com/hnakamur/go-zabbix/internal/rpc"
	"github.com/hnakamur/go-zabbix/internal/slicex"
)

type Host struct {
	HostID            string
	Name              string
	MaintenanceFrom   time.Time
	MaintenanceStatus string
	MaintenanceType   MaintenanceType
	MaintenanceID     string
}

func fromRPCHost(h rpc.Host) (Host, error) {
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

func toRPCHost(h Host) (rpc.Host, error) {
	return rpc.Host{
		HostID: h.HostID,
		Name:   h.Name,
		// Keep empty values for readonly properties
	}, nil
}

func (c *myClient) GetHostsByNamesFullMatch(ctx context.Context,
	names []string) ([]Host, error) {
	rh, err := c.inner.GetHostsByNamesFullMatch(ctx, names)
	if err != nil {
		return nil, err
	}
	return slicex.FailableMap(rh, fromRPCHost)
}

func (c *myClient) GetHostsByGroupIDs(ctx context.Context,
	groupIDs []string) ([]Host, error) {
	rh, err := c.inner.GetHostsByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	return slicex.FailableMap(rh, fromRPCHost)
}
