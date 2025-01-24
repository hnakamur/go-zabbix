package main

import (
	"context"
	"time"

	"github.com/hnakamur/go-zabbix/internal/rpc"
	"github.com/hnakamur/go-zabbix/internal/slicex"
)

// https://www.zabbix.com/documentation/6.0/en/manual/api/reference/trigger/object
type Trigger struct {
	TriggerID   string
	Description string
	Expression  string
	EventName   string
	Comments    string
	Error       string
	LastChange  time.Time
	State       string
	Status      string
	URL         string
	Value       string
	Groups      []HostGroup
	Hosts       []Host
	Items       []Item
}

func (c *myClient) GetTriggers(ctx context.Context, triggerIDs, hostNames, groupNames, descriptions []string) ([]Trigger, error) {
	var hostIDs, groupIDs, itemIDs []string
	if len(hostNames) > 0 {
		hosts, err := c.inner.GetHostsByNamesFullMatch(ctx, hostNames)
		if err != nil {
			return nil, err
		}
		hostIDs = slicex.Map(hosts, func(h rpc.Host) string {
			return h.HostID
		})
	}
	if len(groupNames) > 0 {
		groups, err := c.inner.GetHostGroupsByNamesFullMatch(ctx, groupNames)
		if err != nil {
			return nil, err
		}
		groupIDs = slicex.Map(groups, func(h rpc.HostGroup) string {
			return h.GroupID
		})
	}

	triggers, err := c.inner.GetTriggers(ctx, triggerIDs, hostIDs, groupIDs, itemIDs, descriptions)
	if err != nil {
		return nil, err
	}
	return slicex.FailableMap(triggers, fromPRCTrigger)
}

func (c *myClient) SetTriggersStatus(ctx context.Context, triggerIDs []string, status rpc.TriggerStatus) ([]string, error) {
	var updatedIDs []string

	for _, triggerID := range triggerIDs {
		ids, err := c.inner.SetTriggersStatus(ctx, triggerID, status)
		updatedIDs = append(updatedIDs, ids...)
		if err != nil {
			return updatedIDs, err
		}
	}
	return updatedIDs, nil
}

func (c *myClient) GetTriggerIDs(ctx context.Context, triggerIDs, hostNames, groupNames, descriptions []string) ([]string, error) {
	var hostIDs, groupIDs, itemIDs []string
	if len(hostNames) > 0 {
		hosts, err := c.inner.GetHostsByNamesFullMatch(ctx, hostNames)
		if err != nil {
			return nil, err
		}
		hostIDs = slicex.Map(hosts, func(h rpc.Host) string {
			return h.HostID
		})
	}
	if len(groupNames) > 0 {
		groups, err := c.inner.GetHostGroupsByNamesFullMatch(ctx, groupNames)
		if err != nil {
			return nil, err
		}
		groupIDs = slicex.Map(groups, func(h rpc.HostGroup) string {
			return h.GroupID
		})
	}

	triggerIDs, err := c.inner.GetTriggerIDs(ctx, triggerIDs, hostIDs, groupIDs, itemIDs, descriptions)
	if err != nil {
		return nil, err
	}
	return triggerIDs, nil
}

func fromPRCTrigger(t rpc.Trigger) (Trigger, error) {
	lastChange, err := ParseTimestamp(t.LastChange)
	if err != nil {
		return Trigger{}, err
	}

	hosts, err := slicex.FailableMap(t.Hosts, fromRPCHost)
	if err != nil {
		return Trigger{}, err
	}

	return Trigger{
		TriggerID:   t.TriggerID,
		Description: t.Description,
		Expression:  t.Expression,
		EventName:   t.EventName,
		Comments:    t.Comments,
		Error:       t.Error,
		LastChange:  time.Time(lastChange),
		State:       t.State,
		Status:      t.Status,
		URL:         t.URL,
		Value:       t.Value,
		Groups:      t.Groups,
		Hosts:       hosts,
		Items:       t.Items,
	}, nil
}

func toRPCTrigger(t Trigger) (rpc.Trigger, error) {
	return rpc.Trigger{
		TriggerID:   t.TriggerID,
		Description: t.Description,
		Expression:  t.Expression,
		EventName:   t.EventName,
		Comments:    t.Comments,
		Error:       t.Error,
		Status:      t.Status,
		URL:         t.URL,
		// Keep empty values for readonly properties
	}, nil
}

type displayTrigger struct {
	TriggerID   string           `json:"triggerid,omitempty"`
	Description string           `json:"description,omitempty"`
	Expression  string           `json:"expression,omitempty"`
	EventName   string           `json:"event_name,omitempty"`
	Comments    string           `json:"comments,omitempty"`
	Error       string           `json:"error,omitempty"`
	LastChange  displayTimestamp `json:"lastchange,omitempty"`
	State       string           `json:"state,omitempty"`
	Status      string           `json:"status,omitempty"`
	URL         string           `json:"url,omitempty"`
	Value       string           `json:"value,omitempty"`
	Groups      []HostGroup      `json:"groups"`
	Hosts       []displayHost    `json:"hosts"`
	Items       []Item           `json:"items"`
}

func toDisplayTrigger(t Trigger) displayTrigger {
	return displayTrigger{
		TriggerID:   t.TriggerID,
		Description: t.Description,
		Expression:  t.Expression,
		EventName:   t.EventName,
		Comments:    t.Comments,
		Error:       t.Error,
		LastChange:  displayTimestamp(time.Time(t.LastChange)),
		State:       t.State,
		Status:      t.Status,
		URL:         t.URL,
		Value:       t.Value,
		Groups:      t.Groups,
		Hosts:       slicex.Map(t.Hosts, toDisplayHost),
		Items:       t.Items,
	}
}
