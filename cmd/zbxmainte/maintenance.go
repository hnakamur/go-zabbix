package main

import (
	"context"
	"strings"
	"time"

	"github.com/hnakamur/go-zabbix/internal/rpc"
	"github.com/hnakamur/go-zabbix/internal/slicex"
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

func fromPRCMaintenance(m rpc.Maintenance) (Maintenance, error) {
	activeSince, err := ParseTimestamp(m.ActiveSince)
	if err != nil {
		return Maintenance{}, err
	}
	activeTill, err := ParseTimestamp(m.ActiveTill)
	if err != nil {
		return Maintenance{}, err
	}
	hosts, err := slicex.FailableMap(m.Hosts, fromRPCHost)
	if err != nil {
		return Maintenance{}, err
	}
	timePeriods, err := slicex.FailableMap(m.TimePeriods, fromRPCTimePeriod)
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

func toRPCMaintenance(m Maintenance) (rpc.Maintenance, error) {
	rpcHosts, err := slicex.FailableMap(m.Hosts, toRPCHost)
	if err != nil {
		return rpc.Maintenance{}, err
	}
	rpcTimePeriods, err := slicex.FailableMap(m.TimePeriods, toRPCTimePeriod)
	if err != nil {
		return rpc.Maintenance{}, err
	}

	return rpc.Maintenance{
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

func fromRPCTimePeriod(p rpc.TimePeriod) (TimePeriod, error) {
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

func toRPCTimePeriod(p TimePeriod) (rpc.TimePeriod, error) {
	return rpc.TimePeriod{
		TimeperiodID:   p.TimeperiodID,
		Period:         Seconds(p.Period).String(),
		TimeperiodType: string(p.TimeperiodType),
		StartDate:      Timestamp(p.StartDate).String(),
	}, nil
}

func (c *myClient) GetMaintenances(ctx context.Context) ([]Maintenance, error) {
	rm, err := c.inner.GetMaintenances(ctx)
	if err != nil {
		return nil, err
	}
	return slicex.FailableMap(rm, fromPRCMaintenance)
}

func (c *myClient) GetMaintenanceByID(ctx context.Context, maintenanceID string) (*Maintenance, error) {
	rm, err := c.inner.GetMaintenanceByID(ctx, maintenanceID)
	if err != nil {
		return nil, err
	}
	m, err := fromPRCMaintenance(*rm)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *myClient) GetMaintenanceByNameFullMatch(ctx context.Context, name string) (*Maintenance, error) {
	rm, err := c.inner.GetMaintenanceByNameFullMatch(ctx, name)
	if err != nil {
		return nil, err
	}
	m, err := fromPRCMaintenance(*rm)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *myClient) CreateMaintenance(ctx context.Context, m *Maintenance) error {
	rm, err := toRPCMaintenance(*m)
	if err != nil {
		return err
	}
	if err := c.inner.CreateMaintenance(ctx, &rm); err != nil {
		return err
	}
	m.MaintenaceID = rm.MaintenaceID
	return nil
}

func (c *myClient) UpdateMaintenance(ctx context.Context, m *Maintenance) error {
	rm, err := toRPCMaintenance(*m)
	if err != nil {
		return err
	}
	return c.inner.UpdateMaintenance(ctx, &rm)
}

func (c *myClient) GetMaintenanceIDsByIDs(ctx context.Context, maintenanceIDs []string) ([]string, error) {
	return c.inner.GetMaintenanceIDsByIDs(ctx, maintenanceIDs)
}

func (c *myClient) GetMaintenanceIDsByNamesFullMatch(ctx context.Context, names []string) ([]string, error) {
	return c.inner.GetMaintenanceIDsByNamesFullMatch(ctx, names)
}

func (c *myClient) DeleteMaintenancesByIDs(ctx context.Context, ids []string) (deletedIDs []string, err error) {
	return c.inner.DeleteMaintenancesByIDs(ctx, ids)
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
		Hosts:        slicex.Map(m.Hosts, toDisplayHost),
		TimePeriods:  slicex.Map(m.TimePeriods, toDisplayTimePeriod),
	}
}

func toDisplayHost(h Host) displayHost {
	return displayHost{
		HostID:            h.HostID,
		Name:              h.Name,
		MaintenanceFrom:   displayTimestamp(time.Time(h.MaintenanceFrom)),
		MaintenanceStatus: string(h.MaintenanceStatus),
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
