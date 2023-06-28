package main

import (
	"context"
	"time"

	"github.com/hnakamur/go-zabbix/internal/rpc"
	"github.com/hnakamur/go-zabbix/internal/slicex"
	"golang.org/x/exp/slices"
)

type Host struct {
	HostID            string
	Name              string
	MaintenanceFrom   time.Time
	MaintenanceStatus MaintenanceStatus
	MaintenanceType   MaintenanceType
	MaintenanceID     string
}

type MaintenanceStatus string

const (
	MaintenanceStatusNoMaintenance MaintenanceStatus = "0"
	MaintenanceStatusInEffect      MaintenanceStatus = "1"
)

func fromRPCHost(h rpc.Host) (Host, error) {
	maintenanceFrom, err := ParseTimestamp(h.MaintenanceFrom)
	if err != nil {
		return Host{}, err
	}

	return Host{
		HostID:            h.HostID,
		Name:              h.Name,
		MaintenanceFrom:   time.Time(maintenanceFrom),
		MaintenanceStatus: MaintenanceStatus(h.MaintenanceStatus),
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

func (c *myClient) GetHostsByHostIDs(ctx context.Context,
	hostIDs []string) ([]Host, error) {
	rh, err := c.inner.GetHostsByHostIDs(ctx, hostIDs)
	if err != nil {
		return nil, err
	}
	return slicex.FailableMap(rh, fromRPCHost)
}

func sortHosts(hosts []Host) {
	slices.SortFunc(hosts, func(h1, h2 Host) bool {
		return h1.Name < h2.Name
	})
}

func concatHostsDeDup(hosts ...[]Host) []Host {
	if hosts == nil {
		return nil
	}
	names := make(map[string]struct{})
	var result []Host
	for _, hh := range hosts {
		for _, h := range hh {
			if _, ok := names[h.Name]; !ok {
				result = append(result, h)
				names[h.Name] = struct{}{}
			}
		}
	}
	return result
}

type Hosts []Host

func (hh Hosts) allMaintenanceStatusExpected(expected MaintenanceStatus) bool {
	for _, h := range hh {
		if h.MaintenanceStatus != expected {
			return false
		}
	}
	return true
}
