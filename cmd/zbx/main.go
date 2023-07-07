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
	"github.com/hnakamur/go-zabbix/internal/errlog"
	"github.com/hnakamur/go-zabbix/internal/outlog"
	"github.com/hnakamur/go-zabbix/internal/rpc"
	"github.com/hnakamur/go-zabbix/internal/slicex"
	"github.com/urfave/cli/v2"
)

const timeFormatRFC3339Minute = "2006-01-02T15:04"

func main() {
	if err := run(os.Args); err != nil {
		errlog.Printf("ERROR %s", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	app := &cli.App{
		Name:    "zbx",
		Usage:   "command line tool for Zabbix",
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
				Usage:   "login username",
				EnvVars: []string{"ZBX_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "login password (shows prompt if both of this and token are empty)",
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
			&cli.GenericFlag{
				Name:    "log-flags",
				Value:   &logFlagsValue{flags: log.LstdFlags},
				Usage:   "flags for logger (no prefix it set to empty)",
				EnvVars: []string{"ZBX_LOG_FLAGS"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "mainte",
				Usage: "create, update, or delete maintenance",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "create a maintenance",
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
								Usage:    `active start time of maintenance (default: same as "--start-date")`,
							},
							&cli.TimestampFlag{
								Name:     "active-till",
								Layout:   timeFormatRFC3339Minute,
								Timezone: time.Local,
								Usage:    `active end time of maintenance (default: "--start-date" + "--period")`,
							},
							&cli.TimestampFlag{
								Name:     "start-date",
								Layout:   timeFormatRFC3339Minute,
								Timezone: time.Local,
								Usage:    "start time of maintenance (default: now)",
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
						Usage:  "get maintenances",
						Action: getMaintenancesAction,
					},
					{
						Name:  "update",
						Usage: "update a maintenance",
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
						Usage: "delete maintenance(s)",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:    "id",
								Aliases: []string{"i"},
								Usage:   "target maintenance ID",
							},
							&cli.StringSliceFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "name of maintenance",
							},
						},
						Action: deleteMaintenanceAction,
					},
					{
						Name:  "status",
						Usage: "show host maintenance statuses",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "id",
								Aliases: []string{"i"},
								Usage:   "target maintenance ID",
							},
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "name of maintenance",
							},
							&cli.BoolFlag{
								Name:    "wait",
								Aliases: []string{"w"},
								Usage:   "wait for all hosts to in maintenance effect status",
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
			},
		},
		Before: func(cCtx *cli.Context) error {
			logFlags := cCtx.Generic("log-flags").(*logFlagsValue).flags
			outlog.SetFlags(logFlags)
			outlog.SetOutput(cCtx.App.Writer)
			errlog.SetFlags(logFlags)
			errlog.SetOutput(cCtx.App.ErrWriter)
			return nil
		},
	}

	return app.Run(args)
}

type logFlagsValue struct {
	flags int
}

func (v *logFlagsValue) Set(value string) error {
	flags, err := outlog.ParseLogFlags(value)
	if err != nil {
		return err
	}
	v.flags = flags
	return nil
}

func (v *logFlagsValue) String() string {
	return outlog.LogFlags(v.flags).String()
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

	period := cCtx.Duration("period")
	startDate := cCtx.Timestamp("start-date")
	if startDate == nil {
		now := time.Now().Truncate(time.Minute)
		startDate = &now
	}

	activeSince := cCtx.Timestamp("active-since")
	if activeSince == nil {
		activeSince = startDate
	}

	activeTill := cCtx.Timestamp("active-till")
	if activeTill == nil {
		endDate := startDate.Add(period)
		activeTill = &endDate
	}

	maintenance := &Maintenance{
		Name:           cCtx.String("name"),
		ActiveSince:    *activeSince,
		ActiveTill:     *activeTill,
		Description:    cCtx.String("desc"),
		MaintenaceType: MaintenanceTypeWithData,
		TagsEvalType:   TagsEvalTypeAndOr,
		Groups:         groupsJustID,
		Hosts:          hostsJustID,
		TimePeriods: []TimePeriod{
			{
				Period:         period,
				TimeperiodType: TimeperiodTypeOnetimeOnly,
				StartDate:      *startDate,
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
	outlog.Printf("INFO created maintenance, url: %s", u.String())

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
	outlog.Printf("INFO updated maintenance, url: %s", u.String())

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

	outlog.Printf("INFO maintenance count: %d", len(maintenances))
	for i, m := range maintenances {
		dm := toDisplayMaintenance(m)
		resultBytes, err := json.Marshal(dm)
		if err != nil {
			return err
		}
		outlog.Printf("INFO maintenance i=%d, %s", i, string(resultBytes))
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
	outlog.Printf("INFO targetIDs=%v, deletedIDs=%v", targetIDs, deletedIDs)
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

	{
		maintenanceBytes, err := json.Marshal(toDisplayMaintenance(*maintenance))
		if err != nil {
			panic(err)
		}
		outlog.Printf("INFO maintenance=%s", string(maintenanceBytes))
	}

	logHosts := func(hosts []Host) {
		hostsBytes, err := json.Marshal(slicex.Map(hosts, toDisplayHost))
		if err != nil {
			panic(err)
		}
		outlog.Printf("INFO hosts=%s", string(hostsBytes))

	}
	logHosts(hosts)

	if !cCtx.Bool("wait") {
		return nil
	}

	var hostIDs []string
	interval := cCtx.Duration("interval")
	var timer *time.Timer
	for {
		if Hosts(hosts).allMaintenanceStatusExpected(MaintenanceStatusInEffect) {
			outlog.Printf("INFO all hosts in specified maintenance become in effect status")
			logHosts(hosts)
			return nil
		}

		if timer == nil {
			timer = time.NewTimer(interval)
			defer timer.Stop()
		} else {
			timer.Reset(interval)
		}
		outlog.Print("waiting for maintenance statuses change in all hosts...")
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
