package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

func getHostID(client *zabbix.Client, hostname string) (string, error) {
	type filter struct {
		Host string `json:"host"`
	}
	params := struct {
		Filter filter   `json:"filter"`
		Output []string `json:"output"`
	}{
		Filter: filter{
			Host: hostname,
		},
		Output: []string{"hostid"},
	}

	var hosts []struct {
		HostID string `json:"hostid"`
	}
	err := client.Call("host.get", params, &hosts)
	if err != nil {
		return "", err
	}
	if len(hosts) != 1 {
		return "", fmt.Errorf("expected 1 host but got %d", len(hosts))
	}
	return hosts[0].HostID, nil
}

func getItemID(client *zabbix.Client, itemKey string) (string, error) {
	type search struct {
		Key string `json:"key_"`
	}
	params := struct {
		Search search   `json:"search"`
		Output []string `json:"output"`
	}{
		Search: search{
			Key: itemKey,
		},
		Output: []string{"itemid"},
	}
	var items []struct {
		ItemID string `json:"itemid"`
	}
	err := client.Call("item.get", params, &items)
	if err != nil {
		return "", err
	}
	if len(items) != 1 {
		return "", fmt.Errorf("expected 1 item but got %d", len(items))
	}
	return items[0].ItemID, nil
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

	count, err := getHostCount(client)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("host count=%d", count)

	hostID, err := getHostID(client, "xxxxx.example.com")
	if err != nil {
		logger.Fatal(err)
	}

	itemID, err := getItemID(client, "body_bytes_sent")
	if err != nil {
		logger.Fatal(err)
	}

	params := map[string]interface{}{
		"history":   3,
		"hostids":   hostID,
		"itemids":   itemID,
		"time_from": zabbix.Timestamp(time.Date(2016, 3, 23, 22, 33, 0, 0, time.Local)),
	}
	var histories []struct {
		Clock zabbix.Timestamp `json:"clock,string"`
		Value uint64           `json:"value,string"`
	}
	err = client.Call("history.get", params, &histories)
	if err != nil {
		logger.Fatal(err)
	}
	for i, history := range histories {
		logger.Printf("history[%d]: %v", i, history)
	}
}
