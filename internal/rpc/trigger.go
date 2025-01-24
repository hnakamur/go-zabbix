package rpc

import (
	"context"

	"github.com/hnakamur/go-zabbix/internal/slicex"
)

// https://www.zabbix.com/documentation/6.0/en/manual/api/reference/trigger/object

type Trigger struct {
	TriggerID   string      `json:"triggerid,omitempty"`
	Description string      `json:"description,omitempty"`
	Expression  string      `json:"expression,omitempty"`
	EventName   string      `json:"event_name,omitempty"`
	Comments    string      `json:"comments,omitempty"`
	Error       string      `json:"error,omitempty"`
	LastChange  string      `json:"lastchange,omitempty"`
	State       string      `json:"state,omitempty"`
	Status      string      `json:"status,omitempty"`
	URL         string      `json:"url,omitempty"`
	Value       string      `json:"value,omitempty"`
	Groups      []HostGroup `json:"groups"`
	Hosts       []Host      `json:"hosts"`
	Items       []Item      `json:"items"`
}

func (c *Client) GetTriggers(ctx context.Context, triggerIDs, hostIDs, groupIDs, itemIDs, descriptions []string) ([]Trigger, error) {
	var filter any
	if len(descriptions) > 0 {
		filter = struct {
			Descriptions []string `json:"description"`
		}{
			Descriptions: descriptions,
		}
	}
	params := struct {
		TriggerIDs   []string `json:"triggerids,omitempty"`
		Output       any      `json:"output"`
		Filter       any      `json:"filter,omitempty"`
		HostIDs      any      `json:"hostids,omitempty"`
		GroupIDs     any      `json:"gorupids,omitempty"`
		ItemIDs      any      `json:"itemids,omitempty"`
		SelectGroups any      `json:"selectGroups"`
		SelectHosts  any      `json:"selectHosts"`
		SelectItems  any      `json:"selectItems"`
	}{
		TriggerIDs:   triggerIDs,
		Output:       "extend",
		Filter:       filter,
		HostIDs:      hostIDs,
		GroupIDs:     groupIDs,
		ItemIDs:      itemIDs,
		SelectGroups: selectGroups,
		SelectHosts:  selectHosts,
		SelectItems:  selectItems,
	}
	var triggers []Trigger
	if err := c.Client.Call(ctx, "trigger.get", params, &triggers); err != nil {
		return nil, err
	}
	return triggers, nil
}

func (c *Client) GetTriggerIDs(ctx context.Context, triggerIDs, hostIDs, groupIDs, itemIDs, descriptions []string) ([]string, error) {
	triggers, err := c.GetTriggers(ctx, triggerIDs, hostIDs, groupIDs, itemIDs, descriptions)
	if err != nil {
		return nil, err
	}
	return slicex.Map(triggers, func(t Trigger) string {
		return t.TriggerID
	}), nil
}

type TriggerStatus string

const (
	TriggerStatusEnabled  TriggerStatus = "0"
	TriggerStatusDisabled TriggerStatus = "1"
)

func (c *Client) SetTriggersStatus(ctx context.Context, triggerID string, status TriggerStatus) ([]string, error) {
	type TriggerIDs struct {
		TriggerIDs []string `json:"triggerids"`
	}
	var ids TriggerIDs

	params := struct {
		TriggerID string `json:"triggerid"`
		Status    string `json:"status"`
	}{
		TriggerID: triggerID,
		Status:    string(status),
	}
	if err := c.Client.Call(ctx, "trigger.update", params, &ids); err != nil {
		return nil, err
	}
	return ids.TriggerIDs, nil
}
