package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hnakamur/go-zabbix"
)

func getHostID(client *zabbix.Client, hostname string) (string, error) {
	params := map[string]interface{}{
		"filter": map[string]interface{}{
			"host": hostname,
		},
		"output": []string{"hostid"},
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
	params := map[string]interface{}{
		"search": map[string]interface{}{
			"key_": itemKey,
		},
		"output": []string{"itemid"},
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

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	client := zabbix.NewClient("http://localhost/zabbix")
	client.Logger = logger
	err := client.Login("Admin", "zabbix")
	if err != nil {
		log.Fatal(err)
	}

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
