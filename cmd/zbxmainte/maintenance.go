package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hnakamur/go-zabbix"
)

// https://www.zabbix.com/documentation/6.0/en/manual/api/reference/maintenance/object
type Maintenance struct {
	MaintenaceID   zabbix.ID
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

type MaintenanceType int32

const (
	MaintenanceTypeWithData MaintenanceType = 0
	MaintenanceTypeNoData   MaintenanceType = 1
)

type TagsEvalType int32

const (
	TagsEvalTypeAndOr TagsEvalType = 0
	TagsEvalTypeOr    TagsEvalType = 1
)

type rpcMaintenance struct {
	MaintenaceID   zabbix.ID        `json:"maintenanceid,omitempty"`
	Name           string           `json:"name,omitempty"`
	ActiveSince    zabbix.Timestamp `json:"active_since,omitempty"`
	ActiveTill     zabbix.Timestamp `json:"active_till,omitempty"`
	Description    string           `json:"description,omitempty"`
	MaintenaceType zabbix.Int32     `json:"maintenance_type,omitempty"`
	TagsEvalType   zabbix.Int32     `json:"tags_evaltype,omitempty"`
	Groups         []HostGroup      `json:"groups"`
	Hosts          []Host           `json:"hosts"`
	TimePeriods    []TimePeriod     `json:"timeperiods,omitempty"`
}

func (m Maintenance) MarshalJSON() ([]byte, error) {
	rm := rpcMaintenance{
		MaintenaceID:   m.MaintenaceID,
		Name:           m.Name,
		ActiveSince:    zabbix.Timestamp(m.ActiveSince),
		ActiveTill:     zabbix.Timestamp(m.ActiveTill),
		Description:    m.Description,
		MaintenaceType: zabbix.Int32(m.MaintenaceType),
		TagsEvalType:   zabbix.Int32(m.TagsEvalType),
		Groups:         m.Groups,
		Hosts:          m.Hosts,
		TimePeriods:    m.TimePeriods,
	}
	return json.Marshal(rm)
}

func (m *Maintenance) UnmarshalJSON(data []byte) error {
	var rm rpcMaintenance
	if err := json.Unmarshal(data, &rm); err != nil {
		return err
	}
	*m = Maintenance{
		MaintenaceID:   rm.MaintenaceID,
		Name:           rm.Name,
		ActiveSince:    time.Time(rm.ActiveSince),
		ActiveTill:     time.Time(rm.ActiveTill),
		Description:    rm.Description,
		MaintenaceType: MaintenanceType(rm.MaintenaceType),
		TagsEvalType:   TagsEvalType(rm.TagsEvalType),
		Groups:         rm.Groups,
		Hosts:          rm.Hosts,
		TimePeriods:    rm.TimePeriods,
	}
	return nil
}

type TimePeriodEvery int32

const (
	TimePeriodEveryFirstWeek  TimePeriodEvery = 1
	TimePeriodEverySecondWeek TimePeriodEvery = 2
	TimePeriodEveryThirdWeek  TimePeriodEvery = 3
	TimePeriodEveryFourthWeek TimePeriodEvery = 4
	TimePeriodEveryLastWeek   TimePeriodEvery = 5
)

type TimeperiodType int32

const (
	TimeperiodTypeOnetimeOnly TimeperiodType = 0
	TimeperiodTypeDaily       TimeperiodType = 2
	TimeperiodTypeWeekly      TimeperiodType = 3
	TimeperiodTypeMonthly     TimeperiodType = 4
)

type TimePeriod struct {
	TimeperiodID   zabbix.ID
	Period         time.Duration
	TimeperiodType TimeperiodType
	StartDate      time.Time
	// 期間のタイプが「一度限り」以外の場合の開始時刻の0時からの秒数
	StartTime time.Duration
	Every     TimePeriodEvery
	DayOfWeek int32
	Day       int32
	Month     int32
}

type rpcTimePeriod struct {
	TimeperiodID   zabbix.ID        `json:"timeperiodid,omitempty"`
	Period         zabbix.Int32     `json:"period,omitempty"`
	TimeperiodType zabbix.Int32     `json:"timeperiod_type,omitempty"`
	StartDate      zabbix.Timestamp `json:"start_date,omitempty"`
	StartTime      zabbix.Int32     `json:"start_time,omitempty"`
	Every          zabbix.Int32     `json:"every,omitempty"`
	DayOfWeek      zabbix.Int32     `json:"dayofweek,omitempty"`
	Day            zabbix.Int32     `json:"day,omitempty"`
	Month          zabbix.Int32     `json:"month,omitempty"`
}

func (t TimePeriod) MarshalJSON() ([]byte, error) {
	rt := rpcTimePeriod{
		TimeperiodID:   t.TimeperiodID,
		Period:         zabbix.Int32(t.Period / time.Second),
		TimeperiodType: zabbix.Int32(t.TimeperiodType),
		StartDate:      zabbix.Timestamp(t.StartDate),
		StartTime:      zabbix.Int32(t.StartTime / time.Second),
		Every:          zabbix.Int32(t.Every),
		DayOfWeek:      zabbix.Int32(t.DayOfWeek),
		Day:            zabbix.Int32(t.Day),
		Month:          zabbix.Int32(t.Month),
	}
	return json.Marshal(rt)
}

func (t *TimePeriod) UnmarshalJSON(data []byte) error {
	var rt rpcTimePeriod
	if err := json.Unmarshal(data, &rt); err != nil {
		return err
	}
	*t = TimePeriod{
		TimeperiodID:   rt.TimeperiodID,
		Period:         time.Duration(rt.Period) * time.Second,
		TimeperiodType: TimeperiodType(rt.TimeperiodType),
		StartDate:      time.Time(rt.StartDate),
		StartTime:      time.Duration(rt.StartTime) * time.Second,
		Every:          TimePeriodEvery(rt.Every),
		DayOfWeek:      int32(rt.DayOfWeek),
		Day:            int32(rt.Day),
		Month:          int32(rt.Month),
	}
	return nil
}

func (c *myClient) GetMaintenances(ctx context.Context) ([]Maintenance, error) {
	params := struct {
		Output            any `json:"output"`
		SelectGroups      any `json:"selectGroups"`
		SelectHosts       any `json:"selectHosts"`
		SelectTimeperiods any `json:"selectTimeperiods"`
	}{
		Output:            "extend",
		SelectGroups:      []string{"groupid", "name"},
		SelectHosts:       []string{"hostid", "name"},
		SelectTimeperiods: "extend",
	}
	var result []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	return result, nil
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
		SelectGroups:      []string{"groupid", "name"},
		SelectHosts:       []string{"hostid", "name"},
		SelectTimeperiods: "extend",
		Filter:            Filter{MaintenanceID: []string{maintenanceID}},
	}
	var result []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	if len(result) != 1 {
		return nil, fmt.Errorf("unexpected maintenance count, got=%d, want=1", len(result))
	}
	return &result[0], nil
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
		SelectGroups:      []string{"groupid", "name"},
		SelectHosts:       []string{"hostid", "name"},
		SelectTimeperiods: "extend",
		Filter:            Names{Name: []string{name}},
	}
	var result []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	if len(result) != 1 {
		return nil, fmt.Errorf("unexpected maintenance count, got=%d, want=1", len(result))
	}
	return &result[0], nil
}

func (c *myClient) CreateMaintenance(ctx context.Context, m *Maintenance) error {
	type MaintenanceIDs struct {
		MaintenanceIDs []zabbix.ID `json:"maintenanceids"`
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

func (c *myClient) UpdateMaintenance(ctx context.Context, m *Maintenance) error {
	type MaintenanceIDs struct {
		MaintenanceIDs []zabbix.ID `json:"maintenanceids"`
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

func (c *myClient) GetMaintenanceIDsByIDs(ctx context.Context, maintenanceIDs []string) ([]zabbix.ID, error) {
	type Filter struct {
		MaintenanceID []string `json:"maintenanceid"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: "maintenanceid",
		Filter: Filter{MaintenanceID: maintenanceIDs},
	}
	var result []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	if len(result) != len(maintenanceIDs) {
		return nil, fmt.Errorf("unexpected maintenance count returned by GetMaintenanceIDsByID: got=%d, want=%d", len(result), len(maintenanceIDs))
	}
	ids := MapSlice(result, func(m Maintenance) zabbix.ID {
		return m.MaintenaceID
	})
	return ids, nil
}

func (c *myClient) GetMaintenanceIDsByNamesFullMatch(ctx context.Context, names []string) ([]zabbix.ID, error) {
	type Names struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: "maintenanceid",
		Filter: Names{Name: names},
	}
	var result []Maintenance
	if err := c.Client.Call(ctx, "maintenance.get", params, &result); err != nil {
		return nil, err
	}
	if len(result) != len(names) {
		return nil, fmt.Errorf("unexpected maintenance count returned by GetMaintenanceIDsByNamesFullMatch: got=%d, want=%d", len(result), len(names))
	}
	maintenanceIDs := MapSlice(result, func(m Maintenance) zabbix.ID {
		return m.MaintenaceID
	})
	return maintenanceIDs, nil
}

func (c *myClient) DeleteMaintenancesByIDs(ctx context.Context, ids []zabbix.ID) (deletedIDs []zabbix.ID, err error) {
	type MaintenanceIDs struct {
		MaintenanceIDs []zabbix.ID `json:"maintenanceids"`
	}

	var maintenances MaintenanceIDs
	if err := c.Client.Call(ctx, "maintenance.delete", ids, &maintenances); err != nil {
		return nil, err
	}
	return maintenances.MaintenanceIDs, nil
}

type displayMaintenance struct {
	MaintenaceID zabbix.ID           `json:"maintenanceid,omitempty"`
	Name         string              `json:"name,omitempty"`
	ActiveSince  displayTimestamp    `json:"active_since,omitempty"`
	ActiveTill   displayTimestamp    `json:"active_till,omitempty"`
	Description  string              `json:"description,omitempty"`
	Groups       []HostGroup         `json:"groups"`
	Hosts        []Host              `json:"hosts"`
	TimePeriods  []displayTimePeriod `json:"timeperiods,omitempty"`
}

type displayTimePeriod struct {
	Period    displayDuration
	StartDate displayTimestamp
}

func toDisplayMaintenance(m Maintenance) displayMaintenance {
	return displayMaintenance{
		MaintenaceID: m.MaintenaceID,
		Name:         m.Name,
		ActiveSince:  displayTimestamp(m.ActiveSince),
		ActiveTill:   displayTimestamp(m.ActiveTill),
		Description:  m.Description,
		Groups:       m.Groups,
		Hosts:        m.Hosts,
		TimePeriods:  MapSlice(m.TimePeriods, toDisplayTimePeriod),
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
