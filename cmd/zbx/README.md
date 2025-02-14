# zbx

A command line tool to create, update, or delete a Zabbix maintenance.

## Limitations

- Tested with Zabbix server version 6.0.16.
- Maintenance problem tags are not supported.
- Only supported `timeperiod_type` is "One time only".
- Only one `timeperiod` is supported (multiple `timeperiod`s are not supported).

See the following pages for Maintenance object properties and example.
- https://www.zabbix.com/documentation/6.0/en/manual/api/reference/maintenance/object
- https://www.zabbix.com/documentation/6.0/en/manual/api/reference/maintenance/get

## How to install

### Install binary

An executable file for Linux amd64 can be downloaded from [releases](https://github.com/hnakamur/go-zabbix/releases).

### Install from source

```
go install -trimpath -tags netgo github.com/hnakamur/go-zabbix/cmd/zbx@latest
```

### Usage

1. Set necessary environment variables
   ```
   export ZBX_URL='http://zabbix.example.jp/zabbix'
   export ZBX_API_TOKEN='api_token_generated_by_zabbix'
   ```
1. Learn a little about how to use
   ```
   zbx help
   ```
