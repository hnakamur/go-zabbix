package zabbix

// Logger is an interface for debug logs from go-zabbix library
type Logger interface {
	Log(v interface{})
}
