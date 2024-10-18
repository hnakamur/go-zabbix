package main

import (
	"context"

	"github.com/hnakamur/go-zabbix/internal/rpc"
)

type HostGroup = rpc.HostGroup

func (c *myClient) GetHostGroupsByNamesFullMatch(ctx context.Context,
	names []string) ([]HostGroup, error) {
	return c.inner.GetHostGroupsByNamesFullMatch(ctx, names)
}

func (c *myClient) GetNestedHostGroupsByAncestorNames(ctx context.Context,
	names []string) ([]HostGroup, error) {
	return c.inner.GetNestedHostGroupsByAncestorNames(ctx, names)
}
