package rpc

import (
	"context"
	"fmt"

	"github.com/hnakamur/go-zabbix/internal/slicex"
)

// https://www.zabbix.com/documentation/6.0/en/manual/api/reference/maintenance/object

type Maintenance struct {
	MaintenaceID   string       `json:"maintenanceid,omitempty"`
	Name           string       `json:"name,omitempty"`
	ActiveSince    string       `json:"active_since,omitempty"`
	ActiveTill     string       `json:"active_till,omitempty"`
	Description    string       `json:"description,omitempty"`
	MaintenaceType string       `json:"maintenance_type,omitempty"`
	TagsEvalType   string       `json:"tags_evaltype,omitempty"`
	Groups         []HostGroup  `json:"groups"`
	Hosts          []Host       `json:"hosts"`
	TimePeriods    []TimePeriod `json:"timeperiods,omitempty"`
}

type TimePeriod struct {
	TimeperiodID   string `json:"timeperiodid,omitempty"`
	Period         string `json:"period"`
	TimeperiodType string `json:"timeperiod_type"`
	StartDate      string `json:"start_date"`
}

var selectTimeperiods = []string{"timeperiodid", "period", "timeperiod_type",
	"start_date"}

func (c *Client) GetMaintenances(ctx context.Context) ([]Maintenance, error) {
	params := struct {
		Output            any `json:"output"`
		SelectGroups      any `json:"selectGroups"`
		SelectHosts       any `json:"selectHosts"`
		SelectTimeperiods any `json:"selectTimeperiods"`
	}{
		Output:            "extend",
		SelectGroups:      selectGroups,
		SelectHosts:       selectHosts,
		SelectTimeperiods: selectTimeperiods,
	}
	var rm []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	return rm, nil
}

func (c *Client) GetMaintenanceByID(ctx context.Context, maintenanceID string) (*Maintenance, error) {
	type Filter struct {
		MaintenanceID []string `json:"maintenanceid"`
	}

	params := struct {
		Output            any `json:"output"`
		SelectGroups      any `json:"selectGroups"`
		SelectHosts       any `json:"selectHosts"`
		SelectTimeperiods any `json:"selectTimeperiods"`
		Filter            any `json:"filter"`
	}{
		Output:            "extend",
		SelectGroups:      selectGroups,
		SelectHosts:       selectHosts,
		SelectTimeperiods: selectTimeperiods,
		Filter:            Filter{MaintenanceID: []string{maintenanceID}},
	}
	var rm []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	if len(rm) != 1 {
		return nil, fmt.Errorf("unexpected maintenance count, got=%d, want=1", len(rm))
	}
	return &rm[0], nil
}

func (c *Client) GetMaintenanceByNameFullMatch(ctx context.Context, name string) (*Maintenance, error) {
	type Names struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output            any `json:"output"`
		SelectGroups      any `json:"selectGroups"`
		SelectHosts       any `json:"selectHosts"`
		SelectTimeperiods any `json:"selectTimeperiods"`
		Filter            any `json:"filter"`
	}{
		Output:            "extend",
		SelectGroups:      selectGroups,
		SelectHosts:       selectHosts,
		SelectTimeperiods: selectTimeperiods,
		Filter:            Names{Name: []string{name}},
	}
	var rm []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	if len(rm) != 1 {
		return nil, fmt.Errorf("unexpected maintenance count, got=%d, want=1", len(rm))
	}
	return &rm[0], nil
}

func (c *Client) CreateMaintenance(ctx context.Context, m *Maintenance) error {
	type MaintenanceIDs struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}

	var ids MaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.create", m, &ids); err != nil {
		return err
	}
	if len(ids.MaintenanceIDs) != 1 {
		return fmt.Errorf("unexpected ids length: %d", len(ids.MaintenanceIDs))
	}
	m.MaintenaceID = ids.MaintenanceIDs[0]
	return nil
}

func (c *Client) UpdateMaintenance(ctx context.Context, m *Maintenance) error {
	type MaintenanceIDs struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}

	var ids MaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.update", m, &ids); err != nil {
		return err
	}
	if len(ids.MaintenanceIDs) != 1 {
		return fmt.Errorf("unexpected ids length: %d", len(ids.MaintenanceIDs))
	}
	m.MaintenaceID = ids.MaintenanceIDs[0]
	return nil
}

func (c *Client) GetMaintenanceIDsByIDs(ctx context.Context, maintenanceIDs []string) ([]string, error) {
	type rpcFilter struct {
		MaintenanceID []string `json:"maintenanceid"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: "maintenanceid",
		Filter: rpcFilter{MaintenanceID: maintenanceIDs},
	}
	var rm []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	if len(rm) != len(maintenanceIDs) {
		return nil, fmt.Errorf("unexpected maintenance count returned by GetMaintenanceIDsByID: got=%d, want=%d", len(rm), len(maintenanceIDs))
	}
	ids := slicex.Map(rm, func(m Maintenance) string {
		return m.MaintenaceID
	})
	return ids, nil
}

func (c *Client) GetMaintenanceIDsByNamesFullMatch(ctx context.Context, names []string) ([]string, error) {
	type rpcFilter struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: "maintenanceid",
		Filter: rpcFilter{Name: names},
	}
	var result []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	if len(result) != len(names) {
		return nil, fmt.Errorf("unexpected maintenance count returned by GetMaintenanceIDsByNamesFullMatch: got=%d, want=%d", len(result), len(names))
	}
	maintenanceIDs := slicex.Map(result, func(m Maintenance) string {
		return m.MaintenaceID
	})
	return maintenanceIDs, nil
}

func (c *Client) DeleteMaintenancesByIDs(ctx context.Context, ids []string) (deletedIDs []string, err error) {
	type MaintenanceIDs struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}

	var maintenances MaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.delete", ids, &maintenances); err != nil {
		return nil, err
	}
	return maintenances.MaintenanceIDs, nil
}
