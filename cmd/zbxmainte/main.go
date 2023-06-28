package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime/debug"
	"time"

	"golang.org/x/exp/slices"

	"github.com/hnakamur/go-zabbix"
	"github.com/hnakamur/go-zabbix/internal/rpc"
	"github.com/hnakamur/go-zabbix/internal/slicex"
	"github.com/urfave/cli/v2"
)

const timeFormatRFC3339Minute = "2006-01-02T15:04"

func main() {
	app := &cli.App{
		Name:    "zbxmainte",
		Usage:   "create, get, update, or detele Zabbix maintenance",
		Version: Version(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"l"},
				Usage:    "Zabbix URL (ex. http://example.com/zabbix)",
				Required: true,
				EnvVars:  []string{"ZBX_URL"},
			},
			&cli.StringFlag{
				Name:    "virtual-host",
				Usage:   "virtual host on Zabbix server",
				EnvVars: []string{"ZBX_VIRTUAL_HOST"},
			},
			&cli.StringFlag{
				Name:    "username",
				Aliases: []string{"u"},
				Usage:   "username to login Zabbix",
				EnvVars: []string{"ZBX_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "password to login Zabbix (shows prompt if both of this and token are empty)",
				EnvVars: []string{"ZBX_PASSWORD"},
			},
			&cli.StringFlag{
				Name:    "token",
				Usage:   "Zabbix API token",
				EnvVars: []string{"ZBX_API_TOKEN"},
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "print JSON-RPC requests and responses",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "verbose output",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create Zabbix maintenance",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Required: true,
						Usage:    "name of maintenance to create",
					},
					&cli.StringFlag{
						Name:    "desc",
						Aliases: []string{"d"},
						Usage:   "description of maintenance",
					},
					&cli.StringSliceFlag{
						Name:    "group",
						Aliases: []string{"g"},
						Usage:   "host group names",
					},
					&cli.StringSliceFlag{
						Name:    "host",
						Aliases: []string{"H"},
						Usage:   "host names",
					},
					&cli.TimestampFlag{
						Name:     "active-since",
						Layout:   timeFormatRFC3339Minute,
						Timezone: time.Local,
						Required: true,
						Usage:    "active start time of maintenance",
					},
					&cli.TimestampFlag{
						Name:     "active-till",
						Layout:   timeFormatRFC3339Minute,
						Timezone: time.Local,
						Required: true,
						Usage:    "active end time of maintenance",
					},
					&cli.TimestampFlag{
						Name:     "start-date",
						Layout:   timeFormatRFC3339Minute,
						Timezone: time.Local,
						Required: true,
						Usage:    "start time of maintenance",
					},
					&cli.DurationFlag{
						Name:     "period",
						Required: true,
						Usage:    "duration of maintenance",
					},
				},
				Action: createMaintenanceAction,
			},
			{
				Name:   "get",
				Usage:  "get Zabbix maintenances",
				Action: getMaintenancesAction,
			},
			{
				Name:  "update",
				Usage: "update Zabbix maintenance",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "id",
						Aliases: []string{"i"},
						Usage:   "target maintenance ID (name is used for specifying target if empty)",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   `name of target maintenance, or rename maintenance (when "--id" is set)`,
					},
					&cli.StringFlag{
						Name:    "desc",
						Aliases: []string{"d"},
						Usage:   "description of maintenance",
					},
					&cli.StringSliceFlag{
						Name:    "group",
						Aliases: []string{"g"},
						Usage:   "host group names (or set empty string just once to clear hostgroups)",
					},
					&cli.StringSliceFlag{
						Name:    "host",
						Aliases: []string{"H"},
						Usage:   "host names (or set empty string just once to clear hosts)",
					},
					&cli.TimestampFlag{
						Name:     "active-since",
						Layout:   "2006-01-02T15:04",
						Timezone: time.Local,
						Usage:    "active start time of maintenance",
					},
					&cli.TimestampFlag{
						Name:     "active-till",
						Layout:   "2006-01-02T15:04",
						Timezone: time.Local,
						Usage:    "active end time of maintenance",
					},
					&cli.TimestampFlag{
						Name:     "start-date",
						Layout:   "2006-01-02T15:04",
						Timezone: time.Local,
						Usage:    "start time of maintenance",
					},
					&cli.DurationFlag{
						Name:  "period",
						Usage: "duration of maintenance",
					},
				},
				Action: updateMaintenanceAction,
			},
			{
				Name:  "delete",
				Usage: "delete a Zabbix maintenance(s)",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "id",
						Aliases: []string{"i"},
						Usage:   "target maintenance ID to delete",
					},
					&cli.StringSliceFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "name of maintenance to delete",
					},
				},
				Action: deleteMaintenanceAction,
			},
			{
				Name:  "status",
				Usage: "show host maintenance statuses for a Zabbix maintenance",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "id",
						Aliases: []string{"i"},
						Usage:   "target maintenance ID to delete",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "name of maintenance to delete",
					},
					&cli.StringFlag{
						Name:    "wait",
						Aliases: []string{"w"},
						Usage:   "wait for all hosts maintenance status to be in effect if 1 is set",
					},
					&cli.DurationFlag{
						Name:  "interval",
						Value: 30 * time.Second,
						Usage: "polling interval",
					},
				},
				Action: showStatusAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("ERROR %s", err)
	}
}

func createMaintenanceAction(cCtx *cli.Context) error {
	hostNames := cCtx.StringSlice("host")
	groupNames := cCtx.StringSlice("group")
	if len(hostNames) == 0 && len(groupNames) == 0 {
		return errors.New(`at least "--host" or "--hostgroup" must be set`)
	}

	client, err := newClient(cCtx)
	if err != nil {
		return err
	}

	hostsJustID := []Host{}
	if len(hostNames) > 0 {
		hosts, err := client.GetHostsByNamesFullMatch(cCtx.Context, hostNames)
		if err != nil {
			return err
		}
		hostsJustID = slicex.Map(hosts, func(h Host) Host {
			return Host{HostID: h.HostID}
		})
	}

	groupsJustID := []HostGroup{}
	if len(groupNames) > 0 {
		groups, err := client.GetHostGroupsByNamesFullMatch(cCtx.Context, groupNames)
		if err != nil {
			return err
		}
		groupsJustID = slicex.Map(groups, func(g HostGroup) HostGroup {
			return HostGroup{GroupID: g.GroupID}
		})
	}

	maintenance := &Maintenance{
		Name:           cCtx.String("name"),
		ActiveSince:    *cCtx.Timestamp("active-since"),
		ActiveTill:     *cCtx.Timestamp("active-till"),
		Description:    cCtx.String("desc"),
		MaintenaceType: MaintenanceTypeWithData,
		TagsEvalType:   TagsEvalTypeAndOr,
		Groups:         groupsJustID,
		Hosts:          hostsJustID,
		TimePeriods: []TimePeriod{
			{
				Period:         cCtx.Duration("period"),
				TimeperiodType: TimeperiodTypeOnetimeOnly,
				StartDate:      *cCtx.Timestamp("start-date"),
			},
		},
	}
	if err := client.CreateMaintenance(cCtx.Context, maintenance); err != nil {
		return err
	}

	u, err := maintenanceURL(cCtx, maintenance.MaintenaceID)
	if err != nil {
		return err
	}
	log.Printf("INFO created maintenance, url: %s", u.String())

	return nil
}

func updateMaintenanceAction(cCtx *cli.Context) error {
	hostNames := cCtx.StringSlice("host")
	groupNames := cCtx.StringSlice("group")

	client, err := newClient(cCtx)
	if err != nil {
		return err
	}

	maintenanceID := cCtx.String("id")
	name := cCtx.String("name")

	var maintenance *Maintenance
	if maintenanceID != "" {
		maintenance, err = client.GetMaintenanceByID(cCtx.Context, maintenanceID)
		if err != nil {
			return err
		}

		if name != "" {
			maintenance.Name = name
		}
	} else {
		maintenance, err = client.GetMaintenanceByNameFullMatch(cCtx.Context, name)
		if err != nil {
			return err
		}
	}
	if len(maintenance.TimePeriods) != 1 {
		return fmt.Errorf("unsupported TimePeriod count: got=%d, want=1", len(maintenance.TimePeriods))
	}

	if len(hostNames) > 0 {
		if len(hostNames) == 1 && hostNames[0] == "" {
			maintenance.Hosts = []Host{}
		} else {
			hosts, err := client.GetHostsByNamesFullMatch(cCtx.Context, hostNames)
			if err != nil {
				return err
			}
			maintenance.Hosts = slicex.Map(hosts, func(h Host) Host {
				return Host{HostID: h.HostID}
			})
		}
	} else {
		maintenance.Hosts = slicex.Map(maintenance.Hosts, func(h Host) Host {
			return Host{HostID: h.HostID}
		})
	}

	if len(groupNames) > 0 {
		if len(groupNames) == 1 && groupNames[0] == "" {
			maintenance.Groups = []HostGroup{}
		} else {
			groups, err := client.GetHostGroupsByNamesFullMatch(cCtx.Context, groupNames)
			if err != nil {
				return err
			}
			maintenance.Groups = slicex.Map(groups, func(g HostGroup) HostGroup {
				return HostGroup{GroupID: g.GroupID}
			})
		}
	} else {
		maintenance.Groups = slicex.Map(maintenance.Groups, func(g HostGroup) HostGroup {
			return HostGroup{GroupID: g.GroupID}
		})
	}

	if s := cCtx.String("desc"); s != "" {
		maintenance.Description = s
	}
	if t := cCtx.Timestamp("active-since"); t != nil {
		maintenance.ActiveSince = *t
	}
	if t := cCtx.Timestamp("active-till"); t != nil {
		maintenance.ActiveTill = *t
	}
	if t := cCtx.Timestamp("start-date"); t != nil {
		maintenance.TimePeriods[0].StartDate = *t
	}
	if t := cCtx.Timestamp("start-date"); t != nil {
		maintenance.TimePeriods[0].StartDate = *t
	}
	if d := cCtx.Duration("period"); d != 0 {
		maintenance.TimePeriods[0].Period = d
	}

	if err := client.UpdateMaintenance(cCtx.Context, maintenance); err != nil {
		return err
	}

	u, err := maintenanceURL(cCtx, maintenance.MaintenaceID)
	if err != nil {
		return err
	}
	log.Printf("INFO updated maintenance, url: %s", u.String())

	return nil
}

func getMaintenancesAction(cCtx *cli.Context) error {
	client, err := newClient(cCtx)
	if err != nil {
		return err
	}

	maintenances, err := client.GetMaintenances(cCtx.Context)
	if err != nil {
		return err
	}
	slices.SortFunc(maintenances, func(a, b Maintenance) bool {
		return a.MaintenaceID < b.MaintenaceID
	})

	log.Printf("INFO maintenance count: %d", len(maintenances))
	for i, m := range maintenances {
		dm := toDisplayMaintenance(m)
		resultBytes, err := json.Marshal(dm)
		if err != nil {
			return err
		}
		log.Printf("INFO maintenance i=%d, %s", i, string(resultBytes))
	}
	return nil
}

func deleteMaintenanceAction(cCtx *cli.Context) error {
	ids := cCtx.StringSlice("id")
	names := cCtx.StringSlice("name")
	if len(ids) == 0 && len(names) == 0 {
		return errors.New(`at least one "--name" or "--id" must be set`)
	}
	if slicex.ContainsDup(ids) {
		return errors.New(`duplicated IDs are set with "--id"`)
	}
	if slicex.ContainsDup(names) {
		return errors.New(`duplicated names are set with "--name"`)
	}

	client, err := newClient(cCtx)
	if err != nil {
		return err
	}

	var idsByIDs, idsByNames []string
	if len(ids) > 0 {
		idsByIDs, err = client.GetMaintenanceIDsByIDs(cCtx.Context, ids)
		if err != nil {
			return err
		}
	}
	if len(names) > 0 {
		idsByNames, err = client.GetMaintenanceIDsByNamesFullMatch(cCtx.Context, names)
		if err != nil {
			return err
		}
	}
	targetIDs := slicex.ConcatDeDup(idsByIDs, idsByNames)
	deletedIDs, err := client.DeleteMaintenancesByIDs(cCtx.Context, targetIDs)
	if err != nil {
		return err
	}
	log.Printf("INFO targetIDs=%v, deletedIDs=%v", targetIDs, deletedIDs)
	return nil
}

func showStatusAction(cCtx *cli.Context) error {
	id := cCtx.String("id")
	name := cCtx.String("name")
	if (id == "" && name == "") || (id != "" && name != "") {
		return errors.New(`just one of "--name" or "--id" must be set`)
	}

	client, err := newClient(cCtx)
	if err != nil {
		return err
	}

	var maintenance *Maintenance
	if id != "" {
		maintenance, err = client.GetMaintenanceByID(cCtx.Context, id)
	} else if name != "" {
		maintenance, err = client.GetMaintenanceByNameFullMatch(cCtx.Context, name)
	}
	if err != nil {
		return err
	}

	var hosts []Host
	if len(maintenance.Groups) == 0 {
		hosts = concatHostsDeDup(maintenance.Hosts)
	} else {
		groupIDs := slicex.Map(maintenance.Groups, func(g HostGroup) string {
			return g.GroupID
		})
		hostsInGroups, err := client.GetHostsByGroupIDs(cCtx.Context, groupIDs)
		if err != nil {
			return err
		}
		hosts = concatHostsDeDup(maintenance.Hosts, hostsInGroups)
	}
	sortHosts(hosts)

	log.Printf("INFO maintenance=%+v", maintenance)
	log.Printf("INFO hosts=%+v", hosts)

	waitStatus := cCtx.String("wait")
	if waitStatus == "" {
		return nil
	}

	if waitStatus != string(MaintenanceStatusInEffect) {
		return fmt.Errorf(`"--wait" must be empty or %q`, string(MaintenanceStatusInEffect))
	}

	var hostIDs []string
	interval := cCtx.Duration("interval")
	var timer *time.Timer
	for {
		if Hosts(hosts).allMaintenanceStatusExpected(MaintenanceStatusInEffect) {
			log.Print("yes, maintanence status of all hosts is expected")
			return nil
		}

		if timer == nil {
			timer = time.NewTimer(interval)
			defer timer.Stop()
		} else {
			timer.Reset(interval)
		}
		log.Print("waiting for maintenance statuses change in all hosts...")
		select {
		case <-cCtx.Context.Done():
			return nil
		case <-timer.C:
		}

		if hostIDs == nil {
			hostIDs = slicex.Map(hosts, func(h Host) string {
				return h.HostID
			})
		}
		hosts, err = client.GetHostsByHostIDs(cCtx.Context, hostIDs)
		if err != nil {
			return err
		}
		log.Printf("hosts=%+v", hosts)
	}
}

func newClient(cCtx *cli.Context) (*myClient, error) {
	zabbixURL := cCtx.String("url")
	hostHeader := cCtx.String("virtual-host")

	var opts []zabbix.ClientOpt
	if hostHeader != "" {
		opts = append(opts, zabbix.WithHost(hostHeader))
	}

	token := cCtx.String("token")
	if token != "" {
		opts = append(opts, zabbix.WithAPIToken(token))
	}
	opts = append(opts, zabbix.WithDebug(cCtx.Bool("debug")))

	c, err := zabbix.NewClient(zabbixURL, opts...)
	if err != nil {
		return nil, err
	}

	client := &myClient{inner: &rpc.Client{Client: c}}
	if token == "" {
		if err := login(cCtx, client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func login(cCtx *cli.Context, c *myClient) error {
	username := cCtx.String("username")
	password := cCtx.String("password")

	if username == "" {
		return errors.New(`"--token" or "--username" must be set`)
	}

	if password == "" {
		p, err := readSecret("Enter password for Zabbix:")
		if err != nil {
			return err
		}
		password = string(p)
	}

	if err := c.inner.Login(cCtx.Context, username, password); err != nil {
		return err
	}
	return nil
}

func maintenanceURL(cCtx *cli.Context, maintenanceID string) (*url.URL, error) {
	zabbixURL, err := url.Parse(cCtx.String("url"))
	if err != nil {
		return nil, err
	}

	u := zabbixURL.JoinPath("maintenance.php")
	v := url.Values{}
	v.Add("form", "update")
	v.Add("maintenanceid", maintenanceID)
	u.RawQuery = v.Encode()
	return u, nil
}

func Version() string {
	// This code is copied from
	// https://blog.lufia.org/entry/2020/12/18/002238

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}
