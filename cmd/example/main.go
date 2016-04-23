package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/AlekSi/zabbix-sender"
	"github.com/hnakamur/go-zabbix"
)

func getHostCount(client *zabbix.Client) (int64, error) {
	params := struct {
		CountOutput bool `json:"countOutput"`
	}{
		CountOutput: true,
	}
	return client.CallForCount("host.get", params)
}

const zabbixItemValueTypeNumericUnsigned = 3

func createHostGroupIfNotExists(client *zabbix.Client, hostGroupName string) (hostGroupID string, err error) {
	hostGroupID, err = getHostGroupID(client, hostGroupName)
	if err != nil {
		if err != errNotFound {
			return
		}
		hostGroupID, err = createHostGroup(client, hostGroupName)
	}
	return
}

func createHostIfNotExists(client *zabbix.Client, hostGroupID, hostname string) (hostID string, err error) {
	hostID, err = getHostID(client, hostGroupID, hostname)
	if err != nil {
		if err != errNotFound {
			return
		}
		hostID, err = createHost(client, hostGroupID, hostname)
	}
	return
}

func createIntegerItemIfNotExists(client *zabbix.Client, hostID, itemKey, itemName string) (itemID string, err error) {
	itemID, err = getItemID(client, hostID, itemKey)
	if err != nil {
		if err != errNotFound {
			return
		}
		itemID, err = createIntegerItem(client, hostID, itemKey, itemName)
	}
	return
}

var errNotFound = errors.New("not found")

func getHostGroupID(client *zabbix.Client, hostGroupName string) (hostGroupID string, err error) {
	type filter struct {
		Name string `json:"name"`
	}
	params := struct {
		Filter filter   `json:"filter"`
		Output []string `json:"output"`
	}{
		Filter: filter{
			Name: hostGroupName,
		},
		Output: []string{"groupid"},
	}

	var groups []struct {
		GroupID string `json:"groupid"`
	}
	err = client.Call("hostgroup.get", params, &groups)
	if err != nil {
		return
	}
	if len(groups) == 0 {
		err = errNotFound
		return
	}
	hostGroupID = groups[0].GroupID
	return
}

func createHostGroup(client *zabbix.Client, hostGroupName string) (hostGroupID string, err error) {
	params := struct {
		Name string `json:"name"`
	}{
		Name: hostGroupName,
	}

	var result struct {
		GroupIDs []string `json:"groupids"`
	}
	err = client.Call("hostgroup.create", params, &result)
	if err != nil {
		return
	}
	if len(result.GroupIDs) == 0 {
		err = errNotFound
		return
	}
	hostGroupID = result.GroupIDs[0]
	return
}

func getHostID(client *zabbix.Client, hostGroupID, hostname string) (hostID string, err error) {
	type filter struct {
		Host string `json:"host"`
	}
	params := struct {
		GroupIDs string   `json:"groupids"`
		Filter   filter   `json:"filter"`
		Output   []string `json:"output"`
	}{
		GroupIDs: hostGroupID,
		Filter: filter{
			Host: hostname,
		},
		Output: []string{"hostid"},
	}

	var hosts []struct {
		HostID string `json:"hostid"`
	}
	err = client.Call("host.get", params, &hosts)
	if err != nil {
		return
	}
	if len(hosts) == 0 {
		err = errNotFound
		return
	}
	hostID = hosts[0].HostID
	return
}

func createHost(client *zabbix.Client, hostGroupID, hostname string) (hostID string, err error) {
	type hostInterface struct {
		Type  int    `json:"type"`
		Main  int    `json:"main"`
		UseIP int    `json:"useip"`
		IP    string `json:"ip"`
		Port  string `json:"port"`
		DNS   string `json:"dns"`
	}
	type group struct {
		GroupID string `json:"groupid"`
	}
	params := struct {
		Host       string          `json:"host"`
		Interfaces []hostInterface `json:"interfaces"`
		Groups     []group         `json:"groups"`
	}{
		Host: hostname,
		Interfaces: []hostInterface{
			{
				Type:  1, // Agent
				Main:  1,
				UseIP: 1,
				IP:    "127.0.0.1",
				Port:  "10050",
				DNS:   "",
			},
		},
		Groups: []group{
			{GroupID: hostGroupID},
		},
	}

	var result struct {
		HostIDs []string `json:"hostids"`
	}
	err = client.Call("host.create", params, &result)
	if err != nil {
		return
	}
	if len(result.HostIDs) == 0 {
		err = errNotFound
		return
	}
	hostID = result.HostIDs[0]
	return
}

func getItemID(client *zabbix.Client, hostID, itemKey string) (itemID string, err error) {
	type search struct {
		Key string `json:"key_"`
	}
	params := struct {
		HostIDs string   `json:"hostids"`
		Search  search   `json:"search"`
		Output  []string `json:"output"`
	}{
		HostIDs: hostID,
		Search: search{
			Key: itemKey,
		},
		Output: []string{"itemid"},
	}
	var items []struct {
		ItemID string `json:"itemid"`
	}
	err = client.Call("item.get", params, &items)
	if err != nil {
		return
	}
	if len(items) == 0 {
		err = errNotFound
		return
	}
	itemID = items[0].ItemID
	return
}

func createIntegerItem(client *zabbix.Client, hostID, itemKey, itemName string) (itemID string, err error) {
	params := struct {
		Name      string `json:"name"`
		Key       string `json:"key_"`
		HostID    string `json:"hostid"`
		Type      int    `json:"type"`
		ValueType int    `json:"value_type"`
		Status    int    `json:"status"`
	}{
		Name:      itemName,
		Key:       itemKey,
		HostID:    hostID,
		Type:      2, // Zabbix trapper
		ValueType: zabbixItemValueTypeNumericUnsigned,
		Status:    0, // enabled
	}

	var result struct {
		ItemIDs []string `json:"itemids"`
	}
	err = client.Call("item.create", params, &result)
	if err != nil {
		return
	}
	if len(result.ItemIDs) == 0 {
		err = errNotFound
		return
	}
	itemID = result.ItemIDs[0]
	return
}

func getHistories(client *zabbix.Client, params, histories interface{}) error {
	return client.Call("history.get", params, &histories)
}

func sendItemData(zabbixServerAddr, hostname string, data map[string]interface{}, t time.Time, logger zabbix.Logger) error {
	items := zabbix_sender.MakeDataItems(data, hostname)
	for i := range items {
		items[i].Timestamp = t.Unix()
	}
	addr, err := net.ResolveTCPAddr("tcp", zabbixServerAddr)
	if err != nil {
		return err
	}
	res, err := zabbix_sender.Send(addr, items)
	if err != nil {
		return err
	}
	if logger != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			logger.Log(err.Error())
		}
		logger.Log(fmt.Sprintf("sent data. timeout, time=%s data=%s, res=%s", t.Format(time.RFC3339), string(buf), *res))
	}
	return nil
}

type myLogger struct {
	*log.Logger
}

func (l myLogger) Log(v interface{}) {
	l.Print(v)
}

func main() {
	logger := myLogger{log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)}

	client := zabbix.NewClient("http://localhost/zabbix", "", logger)
	err := client.Login("Admin", "zabbix")
	if err != nil {
		logger.Fatal(err)
	}

	hostGroupID, err := createHostGroupIfNotExists(client, "Go-Zabbix example group")
	if err != nil {
		logger.Fatal(err)
	}
	hostname := "web1.example.com"
	hostID, err := createHostIfNotExists(client, hostGroupID, hostname)
	if err != nil {
		logger.Fatal(err)
	}
	itemKey := "access_count"
	itemName := "Access Count"
	itemID, err := createIntegerItemIfNotExists(client, hostID, itemKey, itemName)
	if err != nil {
		logger.Fatal(err)
	}

	count, err := getHostCount(client)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("host count=%d", count)

	zabbixServerAddr := "localhost:10051"

	data := map[string]interface{}{itemKey: 0}
	t := time.Now()
	err = sendItemData(zabbixServerAddr, hostname, data, t, logger)
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Duration(10) * time.Second)

	data = map[string]interface{}{itemKey: 10}
	t = time.Now()
	err = sendItemData(zabbixServerAddr, hostname, data, t, logger)
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Duration(10) * time.Second)

	data = map[string]interface{}{itemKey: 0}
	t = time.Now()
	err = sendItemData(zabbixServerAddr, hostname, data, t, logger)
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Duration(10) * time.Second)

	params := struct {
		History int    `json:"history"`
		HostIDs string `json:"hostids"`
		ItemIDs string `json:"itemids"`
	}{
		History: zabbixItemValueTypeNumericUnsigned,
		HostIDs: hostID,
		ItemIDs: itemID,
	}
	var histories []struct {
		Clock zabbix.Timestamp `json:"clock,string"`
		Value uint64           `json:"value,string"`
	}
	err = getHistories(client, params, &histories)
	if err != nil {
		logger.Fatal(err)
	}
	for i, history := range histories {
		logger.Printf("history[%d]: %v", i, history)
	}
}
