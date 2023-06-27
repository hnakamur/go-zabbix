package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// https://www.zabbix.com/documentation/6.0/en/manual/api/reference/maintenance/object
type Maintenance struct {
	MaintenaceID   string
	Name           string
	ActiveSince    time.Time
	ActiveTill     time.Time
	Description    string
	MaintenaceType MaintenanceType
	TagsEvalType   TagsEvalType
	Groups         []HostGroup
	Hosts          []Host
	TimePeriods    []TimePeriod
}

type MaintenanceType string

const (
	MaintenanceTypeWithData MaintenanceType = "0"
	MaintenanceTypeNoData   MaintenanceType = "1"
)

type TagsEvalType string

const (
	TagsEvalTypeAndOr TagsEvalType = "0"
	TagsEvalTypeOr    TagsEvalType = "1"
)

type rpcMaintenance struct {
	MaintenaceID   string          `json:"maintenanceid,omitempty"`
	Name           string          `json:"name,omitempty"`
	ActiveSince    string          `json:"active_since,omitempty"`
	ActiveTill     string          `json:"active_till,omitempty"`
	Description    string          `json:"description,omitempty"`
	MaintenaceType string          `json:"maintenance_type,omitempty"`
	TagsEvalType   string          `json:"tags_evaltype,omitempty"`
	Groups         []HostGroup     `json:"groups"`
	Hosts          []rpcHost       `json:"hosts"`
	TimePeriods    []rpcTimePeriod `json:"timeperiods,omitempty"`
}

func toMaintenance(m rpcMaintenance) (Maintenance, error) {
	activeSince, err := ParseTimestamp(m.ActiveSince)
	if err != nil {
		return Maintenance{}, err
	}
	activeTill, err := ParseTimestamp(m.ActiveTill)
	if err != nil {
		return Maintenance{}, err
	}
	hosts, err := FailableMapSlice(m.Hosts, toHost)
	if err != nil {
		return Maintenance{}, err
	}
	timePeriods, err := FailableMapSlice(m.TimePeriods, toTimePeriod)
	if err != nil {
		return Maintenance{}, err
	}

	return Maintenance{
		MaintenaceID:   m.MaintenaceID,
		Name:           m.Name,
		ActiveSince:    time.Time(activeSince),
		ActiveTill:     time.Time(activeTill),
		Description:    m.Description,
		MaintenaceType: MaintenanceType(m.MaintenaceType),
		TagsEvalType:   TagsEvalType(m.TagsEvalType),
		Groups:         m.Groups,
		Hosts:          hosts,
		TimePeriods:    timePeriods,
	}, nil
}

func toRPCMaintenance(m Maintenance) (rpcMaintenance, error) {
	rpcHosts, err := FailableMapSlice(m.Hosts, toRPCHost)
	if err != nil {
		return rpcMaintenance{}, err
	}
	rpcTimePeriods, err := FailableMapSlice(m.TimePeriods, toRPCTimePeriod)
	if err != nil {
		return rpcMaintenance{}, err
	}

	return rpcMaintenance{
		MaintenaceID:   m.MaintenaceID,
		Name:           m.Name,
		ActiveSince:    Timestamp(m.ActiveSince).String(),
		ActiveTill:     Timestamp(m.ActiveTill).String(),
		Description:    m.Description,
		MaintenaceType: string(m.MaintenaceType),
		TagsEvalType:   string(m.TagsEvalType),
		Groups:         m.Groups,
		Hosts:          rpcHosts,
		TimePeriods:    rpcTimePeriods,
	}, nil
}

type TimeperiodType string

const (
	TimeperiodTypeOnetimeOnly TimeperiodType = "0"
	TimeperiodTypeDaily       TimeperiodType = "2"
	TimeperiodTypeWeekly      TimeperiodType = "3"
	TimeperiodTypeMonthly     TimeperiodType = "4"
)

type TimePeriod struct {
	TimeperiodID   string
	Period         time.Duration
	TimeperiodType TimeperiodType
	StartDate      time.Time
}

type rpcTimePeriod struct {
	TimeperiodID   string `json:"timeperiodid,omitempty"`
	Period         string `json:"period"`
	TimeperiodType string `json:"timeperiod_type"`
	StartDate      string `json:"start_date"`
}

func toTimePeriod(p rpcTimePeriod) (TimePeriod, error) {
	period, err := ParseSeconds(p.Period)
	if err != nil {
		return TimePeriod{}, err
	}
	startDate, err := ParseTimestamp(p.StartDate)
	if err != nil {
		return TimePeriod{}, err
	}

	return TimePeriod{
		TimeperiodID:   p.TimeperiodID,
		Period:         time.Duration(period),
		TimeperiodType: TimeperiodType(p.TimeperiodType),
		StartDate:      time.Time(startDate),
	}, nil
}

func toRPCTimePeriod(p TimePeriod) (rpcTimePeriod, error) {
	return rpcTimePeriod{
		TimeperiodID:   p.TimeperiodID,
		Period:         Seconds(p.Period).String(),
		TimeperiodType: string(p.TimeperiodType),
		StartDate:      Timestamp(p.StartDate).String(),
	}, nil
}

var selectGroups = []string{"groupid", "name"}
var selectHosts = []string{"hostid", "name", "maintenance_from",
	"maintenance_status", "maintenance_type", "maintenanceid"}
var selectTimeperiods = []string{"timeperiodid", "period", "timeperiod_type",
	"start_date"}

func (c *myClient) GetMaintenances(ctx context.Context) ([]Maintenance, error) {
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
	var rm []rpcMaintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	return FailableMapSlice(rm, toMaintenance)
}

func (c *myClient) GetMaintenanceByID(ctx context.Context, maintenanceID string) (*Maintenance, error) {
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
	var rm []rpcMaintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	if len(rm) != 1 {
		return nil, fmt.Errorf("unexpected maintenance count, got=%d, want=1", len(rm))
	}
	m, err := toMaintenance(rm[0])
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *myClient) GetMaintenanceByNameFullMatch(ctx context.Context, name string) (*Maintenance, error) {
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
	var rm []rpcMaintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	if len(rm) != 1 {
		return nil, fmt.Errorf("unexpected maintenance count, got=%d, want=1", len(rm))
	}
	m, err := toMaintenance(rm[0])
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *myClient) CreateMaintenance(ctx context.Context, m *Maintenance) error {
	type rpcMaintenanceIDs struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}

	var ids rpcMaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.create", m, &ids); err != nil {
		return err
	}
	if len(ids.MaintenanceIDs) != 1 {
		return fmt.Errorf("unexpected ids length: %d", len(ids.MaintenanceIDs))
	}
	m.MaintenaceID = ids.MaintenanceIDs[0]
	return nil
}

func (c *myClient) UpdateMaintenance(ctx context.Context, m *Maintenance) error {
	type rpcMaintenanceIDs struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}

	rm, err := toRPCMaintenance(*m)
	if err != nil {
		return err
	}
	var ids rpcMaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.update", rm, &ids); err != nil {
		return err
	}
	if len(ids.MaintenanceIDs) != 1 {
		return fmt.Errorf("unexpected ids length: %d", len(ids.MaintenanceIDs))
	}
	m.MaintenaceID = ids.MaintenanceIDs[0]
	return nil
}

func (c *myClient) GetMaintenanceIDsByIDs(ctx context.Context, maintenanceIDs []string) ([]string, error) {
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
	var rm []rpcMaintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &rm); err != nil {
		return nil, err
	}
	if len(rm) != len(maintenanceIDs) {
		return nil, fmt.Errorf("unexpected maintenance count returned by GetMaintenanceIDsByID: got=%d, want=%d", len(rm), len(maintenanceIDs))
	}
	ids := MapSlice(rm, func(m rpcMaintenance) string {
		return m.MaintenaceID
	})
	return ids, nil
}

func (c *myClient) GetMaintenanceIDsByNamesFullMatch(ctx context.Context, names []string) ([]string, error) {
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
	var result []rpcMaintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	if len(result) != len(names) {
		return nil, fmt.Errorf("unexpected maintenance count returned by GetMaintenanceIDsByNamesFullMatch: got=%d, want=%d", len(result), len(names))
	}
	maintenanceIDs := MapSlice(result, func(m rpcMaintenance) string {
		return m.MaintenaceID
	})
	return maintenanceIDs, nil
}

func (c *myClient) DeleteMaintenancesByIDs(ctx context.Context, ids []string) (deletedIDs []string, err error) {
	type rpcMaintenanceIDs struct {
		MaintenanceIDs []string `json:"maintenanceids"`
	}

	var maintenances rpcMaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.delete", ids, &maintenances); err != nil {
		return nil, err
	}
	return maintenances.MaintenanceIDs, nil
}

type displayMaintenance struct {
	MaintenaceID string              `json:"maintenanceid"`
	Name         string              `json:"name"`
	ActiveSince  displayTimestamp    `json:"active_since"`
	ActiveTill   displayTimestamp    `json:"active_till"`
	Description  string              `json:"description"`
	Groups       []HostGroup         `json:"groups"`
	Hosts        []displayHost       `json:"hosts"`
	TimePeriods  []displayTimePeriod `json:"timeperiods"`
}

type displayHost struct {
	HostID            string           `json:"hostid"`
	Name              string           `json:"name"`
	MaintenanceFrom   displayTimestamp `json:"maintenance_from"`
	MaintenanceStatus string           `json:"maintenance_status"`
	MaintenanceType   MaintenanceType  `json:"maintenance_type"`
	MaintenanceID     string           `json:"maintenanceid"`
}

type displayTimePeriod struct {
	Period    displayDuration  `json:"period"`
	StartDate displayTimestamp `json:"start_date"`
}

func toDisplayMaintenance(m Maintenance) displayMaintenance {
	return displayMaintenance{
		MaintenaceID: m.MaintenaceID,
		Name:         m.Name,
		ActiveSince:  displayTimestamp(m.ActiveSince),
		ActiveTill:   displayTimestamp(m.ActiveTill),
		Description:  m.Description,
		Groups:       m.Groups,
		Hosts:        MapSlice(m.Hosts, toDisplayHost),
		TimePeriods:  MapSlice(m.TimePeriods, toDisplayTimePeriod),
	}
}

func toDisplayHost(h Host) displayHost {
	return displayHost{
		HostID:            h.HostID,
		Name:              h.Name,
		MaintenanceFrom:   displayTimestamp(time.Time(h.MaintenanceFrom)),
		MaintenanceStatus: h.MaintenanceStatus,
		MaintenanceType:   MaintenanceType(h.MaintenanceType),
		MaintenanceID:     h.MaintenanceID,
	}
}

func toDisplayTimePeriod(tp TimePeriod) displayTimePeriod {
	return displayTimePeriod{
		Period:    displayDuration(tp.Period),
		StartDate: displayTimestamp(tp.StartDate),
	}
}

// Timestamp is an alias to time.Time. Timestamp is encoded a string whose value
// is seconds from the Unix epoch time.
type displayTimestamp time.Time

// MarshalJSON returns a string of seconds from the Unix Epoch time.
func (t displayTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

// UnmarshalJSON reads a string of seconds from the Unix Epoch time.
func (t *displayTimestamp) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.Parse(s, timeFormatRFC3339Minute)
	if err != nil {
		return err
	}
	*t = displayTimestamp(parsed)
	return nil
}

func (t displayTimestamp) String() string {
	return time.Time(t).Format(timeFormatRFC3339Minute)
}

type displayDuration time.Duration

func (d displayDuration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *displayDuration) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = displayDuration(parsed)
	return nil
}

func (d displayDuration) String() string {
	return time.Duration(d).String()
}
