go-zabbix
=========

A minimal Zabbix API client for Go.
See https://www.zabbix.com/documentation/2.4/manual/api for the Zabbix API.

NOTE: I will NOT add functions or types for each Zabbix API.
You can define functions or types yourself on your specific needs in your program.

See cmd/example/main.go for an example.
It uses [Anonymous structs](https://talks.golang.org/2012/10things.slide#3) for the result values.

The timestamp datatype in JSON requests must be encoded in seconds from the Unix epoch time.
You can use zabbix.Timestamp like cmd/example/main.go.

## License
MIT
