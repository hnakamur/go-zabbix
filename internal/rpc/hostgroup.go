package rpc

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type HostGroup struct {
	GroupID string `json:"groupid"`
	Name    string `json:"name,omitempty"`
}

var selectGroups = []string{"groupid", "name"}

func (c *Client) GetHostGroupsByNamesFullMatch(ctx context.Context,
	names []string) ([]HostGroup, error) {
	type Names struct {
		Name []string `json:"name"`
	}

	params := struct {
		Output any `json:"output"`
		Filter any `json:"filter"`
	}{
		Output: []string{"groupid", "name"},
		Filter: Names{
			Name: names,
		},
	}
	var groups []HostGroup
	if err := c.Client.Call(ctx, "hostgroup.get", params, &groups); err != nil {
		return nil, err
	}

	var notFoundNames []string
	for _, name := range names {
		if !slices.ContainsFunc(groups, func(grp HostGroup) bool {
			return grp.Name == name
		}) {
			notFoundNames = append(notFoundNames, name)
		}
	}
	if len(notFoundNames) > 0 {
		return nil, fmt.Errorf("host groups not found: %s", strings.Join(notFoundNames, ", "))
	}
	return groups, nil
}

func (c *Client) GetNestedHostGroupsByAncestorNames(ctx context.Context,
	names []string) ([]HostGroup, error) {

	params := struct {
		Output any `json:"output"`
	}{
		Output: []string{"groupid", "name"},
	}
	var groups []HostGroup
	if err := c.Client.Call(ctx, "hostgroup.get", params, &groups); err != nil {
		return nil, err
	}

	var notFoundNames []string
	for _, name := range names {
		if !slices.ContainsFunc(groups, func(grp HostGroup) bool {
			return grp.Name == name
		}) {
			notFoundNames = append(notFoundNames, name)
		}
	}
	if len(notFoundNames) > 0 {
		return nil, fmt.Errorf("host groups not found: %s", strings.Join(notFoundNames, ", "))
	}
	return filterHostGroupsByAncestorNames(groups, names), nil
}

func filterHostGroupsByAncestorNames(groups []HostGroup, names []string) []HostGroup {
	var filteredGroups []HostGroup
	for _, group := range groups {
		for _, name := range names {
			if group.Name == name || strings.HasPrefix(group.Name, name+"/") {
				filteredGroups = append(filteredGroups, group)
			}
		}
	}
	return filteredGroups
}
