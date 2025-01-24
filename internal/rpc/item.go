package rpc

// https://www.zabbix.com/documentation/6.0/en/manual/api/reference/item/object

type Item struct {
	ItemID string `json:"itemid"`
	HostID string `json:"hostid"`
	Key    string `json:"key_"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

var selectItems = []string{"itemid", "hostid", "key_", "name", "type"}
