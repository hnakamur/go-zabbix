package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/hnakamur/go-zabbix"
)

type myClient struct {
	*zabbix.Client
}

func (c *myClient) getHostCount(ctx context.Context) (int64, error) {
	params := struct {
		CountOutput bool `json:"countOutput"`
	}{
		CountOutput: true,
	}
	var countStr string
	if err := c.Client.Call(ctx, "host.get", params, &countStr); err != nil {
		return 0, err
	}
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func main() {
	zabbixURL := flag.String("zabbix-url", "", "Zabbix URL (ex. http://example.com/zabbix)")
	username := flag.String("username", "", "user name")
	host := flag.String("host", "", "host header for selecting virtualhost")
	flag.Parse()

	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("password must be set with \"PASSWORD\" environment variable")
	}

	var opts []zabbix.ClientOpt
	if *host != "" {
		opts = append(opts, zabbix.WithHost(*host))
	}
	c, err := zabbix.NewClient(*zabbixURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	client := myClient{Client: c}

	ctx := context.Background()
	if err := client.Login(ctx, *username, password); err != nil {
		log.Fatal(err)
	}

	count, err := client.getHostCount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("host count=%d", count)
}
