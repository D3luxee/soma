package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/mjolnir42/soma/internal/adm"
	"github.com/mjolnir42/soma/internal/cmpl"
	"github.com/mjolnir42/soma/lib/proto"
)

func registerDatacenters(app cli.App) *cli.App {
	app.Commands = append(app.Commands,
		[]cli.Command{
			// datacenters
			{
				Name:  "datacenters",
				Usage: "SUBCOMMANDS for datacenters",
				Subcommands: []cli.Command{
					{
						Name:   "add",
						Usage:  "Register a new datacenter",
						Action: runtime(cmdDatacentersAdd),
					},
					{
						Name:   "remove",
						Usage:  "Remove an existing datacenter",
						Action: runtime(cmdDatacentersRemove),
					},
					{
						Name:         "rename",
						Usage:        "Rename an existing datacenter",
						Action:       runtime(cmdDatacentersRename),
						BashComplete: cmpl.To,
					},
					{
						Name:   "list",
						Usage:  "List all datacenters",
						Action: runtime(cmdDatacentersList),
					},
					{
						Name:   "show",
						Usage:  "Show information about a specific datacenter",
						Action: runtime(cmdDatacentersShow),
					},
					{
						Name:   "synclist",
						Usage:  "List all datacenters suitable for sync",
						Action: runtime(cmdDatacentersSync),
					},
				},
			}, // end datacenters
		}...,
	)
	return &app
}

func cmdDatacentersAdd(c *cli.Context) error {
	if err := adm.VerifySingleArgument(c); err != nil {
		return err
	}

	req := proto.NewDatacenterRequest()
	req.Datacenter.LoCode = c.Args().First()

	return adm.Perform(`postbody`, `/datacenter/`, `command`, req, c)
}

func cmdDatacentersRemove(c *cli.Context) error {
	if err := adm.VerifySingleArgument(c); err != nil {
		return err
	}

	path := fmt.Sprintf("/datacenter/%s", c.Args().First())
	return adm.Perform(`delete`, path, `command`, nil, c)
}

func cmdDatacentersRename(c *cli.Context) error {
	key := []string{`to`}
	opts := map[string][]string{}

	if err := adm.ParseVariadicArguments(opts, key, key, key,
		c.Args().Tail()); err != nil {
		return err
	}

	req := proto.NewDatacenterRequest()
	req.Datacenter.LoCode = opts[`to`][0]

	path := fmt.Sprintf("/datacenter/%s", c.Args().First())
	return adm.Perform(`put`, path, `command`, nil, c)
}

func cmdDatacentersList(c *cli.Context) error {
	if err := adm.VerifyNoArgument(c); err != nil {
		return err
	}

	return adm.Perform(`get`, `/datacenter/`, `list`, nil, c)
}

func cmdDatacentersSync(c *cli.Context) error {
	if err := adm.VerifyNoArgument(c); err != nil {
		return err
	}

	return adm.Perform(`get`, `/sync/datacenter/`, `list`, nil, c)
}

func cmdDatacentersShow(c *cli.Context) error {
	if err := adm.VerifySingleArgument(c); err != nil {
		return err
	}

	path := fmt.Sprintf("/datacenter/%s", c.Args().First())
	return adm.Perform(`get`, path, `show`, nil, c)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix