/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mjolnir42/soma/internal/adm"
	"github.com/mjolnir42/soma/internal/cmpl"
	"github.com/mjolnir42/soma/internal/help"
	"github.com/mjolnir42/soma/lib/proto"
)

func registerPermissions(app cli.App) *cli.App {
	app.Commands = append(app.Commands,
		[]cli.Command{
			{
				Name:  "permissions",
				Usage: "SUBCOMMANDS for permissions",
				Subcommands: []cli.Command{
					{
						Name:         "add",
						Usage:        "Register a new permission",
						Description:  help.Text(`PermissionsAdd`),
						Action:       runtime(cmdPermissionAdd),
						BashComplete: cmpl.To,
					},
					{
						Name:         "remove",
						Usage:        "Remove a permission from a category",
						Description:  help.Text(`PermissionsRemove`),
						Action:       runtime(cmdPermissionDel),
						BashComplete: cmpl.From,
					},
					{
						Name:         "list",
						Usage:        "List all permissions in a category",
						Description:  help.Text(`PermissionsList`),
						Action:       runtime(cmdPermissionList),
						BashComplete: cmpl.DirectIn,
					},
					{
						Name:         "show",
						Usage:        "Show details for a permission",
						Description:  help.Text(`PermissionsShow`),
						Action:       runtime(cmdPermissionShow),
						BashComplete: cmpl.In,
					},
					{
						Name:         `map`,
						Usage:        `Map an action to a permission`,
						Description:  help.Text(`PermissionsMap`),
						Action:       runtime(cmdPermissionMap),
						BashComplete: cmpl.To,
					},
					{
						Name:         `unmap`,
						Usage:        `Unmap an action from a permission`,
						Description:  help.Text(`PermissionsUnmap`),
						Action:       runtime(cmdPermissionUnmap),
						BashComplete: cmpl.From,
					},
				},
			}, // end permissions
		}...,
	)
	return &app
}

func cmdPermissionAdd(c *cli.Context) error {
	unique := []string{`to`}
	required := []string{`to`}
	opts := map[string][]string{}
	if err := adm.ParseVariadicArguments(
		opts,
		[]string{},
		unique,
		required,
		c.Args().Tail(),
	); err != nil {
		return err
	}
	if err := adm.ValidateNoColon(c.Args().First()); err != nil {
		return err
	}
	if err := adm.ValidateCategory(opts[`to`][0]); err != nil {
		return err
	}

	esc := url.QueryEscape(opts[`to`][0])
	req := proto.NewPermissionRequest()
	req.Permission.Name = c.Args().First()
	req.Permission.Category = opts[`to`][0]
	path := fmt.Sprintf("/category/%s/permissions/", esc)
	return adm.Perform(`postbody`, path, `command`, req, c)
}

func cmdPermissionDel(c *cli.Context) error {
	unique := []string{`from`}
	required := []string{`from`}
	opts := map[string][]string{}
	if err := adm.ParseVariadicArguments(
		opts,
		[]string{},
		unique,
		required,
		c.Args().Tail(),
	); err != nil {
		return err
	}
	var permissionID string
	if err := adm.ValidateCategory(opts[`from`][0]); err != nil {
		return err
	}
	if err := adm.LookupPermIDRef(c.Args().First(),
		opts[`from`][0],
		&permissionID,
	); err != nil {
		return err
	}

	esc := url.QueryEscape(opts[`from`][0])
	path := fmt.Sprintf("/category/%s/permissions/%s",
		esc, permissionID)
	return adm.Perform(`delete`, path, `command`, nil, c)
}

func cmdPermissionList(c *cli.Context) error {
	unique := []string{`in`}
	required := []string{`in`}
	opts := map[string][]string{}
	if err := adm.ParseVariadicArguments(
		opts,
		[]string{},
		unique,
		required,
		adm.AllArguments(c),
	); err != nil {
		return err
	}
	if err := adm.ValidateCategory(opts[`in`][0]); err != nil {
		return err
	}

	esc := url.QueryEscape(opts[`in`][0])
	path := fmt.Sprintf("/category/%s/permissions/", esc)
	return adm.Perform(`get`, path, `list`, nil, c)
}

func cmdPermissionShow(c *cli.Context) error {
	unique := []string{`in`}
	required := []string{`in`}
	opts := map[string][]string{}
	if err := adm.ParseVariadicArguments(
		opts,
		[]string{},
		unique,
		required,
		c.Args().Tail(),
	); err != nil {
		return err
	}
	var permissionID string
	if err := adm.ValidateCategory(opts[`in`][0]); err != nil {
		return err
	}
	if err := adm.LookupPermIDRef(c.Args().First(),
		opts[`in`][0],
		&permissionID,
	); err != nil {
		return err
	}

	esc := url.QueryEscape(opts[`in`][0])
	path := fmt.Sprintf("/category/%s/permissions/%s",
		esc, permissionID)
	return adm.Perform(`get`, path, `show`, nil, c)
}

func cmdPermissionMap(c *cli.Context) error {
	return cmdPermissionEdit(c, `map`)
}

func cmdPermissionUnmap(c *cli.Context) error {
	return cmdPermissionEdit(c, `unmap`)
}

func cmdPermissionEdit(c *cli.Context, cmd string) error {
	var syn string
	switch cmd {
	case `map`:
		syn = `to`
	case `unmap`:
		syn = `from`
	}
	unique := []string{syn}
	required := []string{syn}
	opts := map[string][]string{}
	if err := adm.ParseVariadicArguments(
		opts,
		[]string{},
		unique,
		required,
		c.Args().Tail(),
	); err != nil {
		return err
	}
	var (
		err                                   error
		section, action, category, permission string
		sectionID, actionID, permissionID     string
		sectionMapping                        bool
	)
	actionSlice := strings.Split(c.Args().First(), `::`)
	permissionSlice := strings.Split(opts[syn][0], `::`)
	switch len(actionSlice) {
	case 1:
		section = actionSlice[0]
		sectionMapping = true
	case 2:
		section = actionSlice[0]
		action = actionSlice[1]
	default:
		return fmt.Errorf("Not a valid {section}::{action}"+
			" specifier: %s", c.Args().First())
	}
	switch len(permissionSlice) {
	case 2:
		category = permissionSlice[0]
		permission = permissionSlice[1]
	default:
		return fmt.Errorf("Not a valid {category}::{permission}"+
			" specifier: %s", opts[syn][0])
	}
	// validate category
	if err = adm.ValidateCategory(category); err != nil {
		return err
	}
	// lookup permissionid
	if err = adm.LookupPermIDRef(
		permission,
		category,
		&permissionID,
	); err != nil {
		return err
	}
	// lookup sectionid
	if sectionID, err = adm.LookupSectionID(
		section,
	); err != nil {
		return err
	}
	if !sectionMapping {
		// lookup actionid
		if actionID, err = adm.LookupActionID(
			action,
			sectionID,
		); err != nil {
			return err
		}
	}

	req := proto.NewPermissionRequest()
	switch cmd {
	case `map`:
		req.Flags.Add = true
	case `unmap`:
		req.Flags.Remove = true
	}
	req.Permission.ID = permissionID
	req.Permission.Name = permission
	req.Permission.Category = category
	if !sectionMapping {
		req.Permission.Actions = &[]proto.Action{
			proto.Action{
				ID:        actionID,
				Name:      action,
				SectionID: sectionID,
				Category:  category,
			},
		}
	} else {
		req.Permission.Sections = &[]proto.Section{
			proto.Section{
				ID:       sectionID,
				Name:     section,
				Category: category,
			},
		}
	}

	esc := url.QueryEscape(category)
	path := fmt.Sprintf("/category/%s/permissions/%s",
		esc, permissionID)
	return adm.Perform(`patchbody`, path, `command`, req, c)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
