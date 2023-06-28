package main

import "github.com/hnakamur/go-zabbix/internal/rpc"

type myClient struct {
	inner *rpc.Client
}
