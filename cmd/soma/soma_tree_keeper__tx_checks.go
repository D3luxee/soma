/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */
package main

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/1and1/soma/lib/proto"
)

func (tk *treeKeeper) txCheckConfig(conf proto.CheckConfig,
	stm *map[string]*sql.Stmt) error {
	if conf.BucketId != "" {
		nullBucket = sql.NullString{
			String: conf.BucketId,
			Valid:  true,
		}
	} else {
		nullBucket = sql.NullString{String: "", Valid: false}
	}
	if _, err = stm[`CreateCheckConfigurationBase`].Exec(
		conf.Id,
		conf.Name,
		int64(conf.Interval),
		conf.RepositoryId,
		nullbucket,
		conf.CapabilityId,
		conf.ObjectId,
		conf.ObjectType,
		conf.IsActive,
		conf.IsEnabled,
		conf.Inheritance,
		conf.ChildrenOnly,
		conf.ExternalId,
	); err != nil {
		return err
	}

threshloop:
	for _, thr := range conf.Thresholds {
		if _, err = stm[`CreateCheckConfigurationThreshold`].Exec(
			conf.Id,
			thr.Predicate.Symbol,
			strconv.FormatInt(thr.Value, 10),
			thr.Level.Name,
		); err != nil {
			break threshloop
		}
	}
	if err != nil {
		return err
	}

constrloop:
	for _, constr := range conf.Constraints {
		switch constr.ConstraintType {
		case "native":
			if _, err = stm[`CreateCheckConfigurationConstraintNative`].Exec(
				conf.Id,
				constr.Native.Name,
				constr.Native.Value,
			); err != nil {
				break constrloop
			}
		case "oncall":
			if _, err = stm[`CreateCheckConfigurationConstraintOncall`].Exec(
				conf.Id,
				constr.Oncall.Id,
			); err != nil {
				break constrloop
			}
		case "custom":
			if _, err = stm[`CreateCheckConfigurationConstraintCustom`].Exec(
				conf.Id,
				constr.Custom.Id,
				constr.Custom.RepositoryId,
				constr.Custom.Value,
			); err != nil {
				break constrloop
			}
		case "system":
			if _, err = stm[`CreateCheckConfigurationConstraintSystem`].Exec(
				conf.Id,
				constr.System.Name,
				constr.System.Value,
			); err != nil {
				break constrloop
			}
		case "service":
			if constr.Service.TeamId != tk.team {
				err = fmt.Errorf(
					"Service constraint has mismatched TeamID values: %s/%s",
					tk.team, constr.Service.TeamId)
				break constrloop
			}
			if _, err = stm[`CreateCheckConfigurationConstraintService`].Exec(
				conf.Id,
				tk.team,
				constr.Service.Name,
			); err != nil {
				break constrloop
			}
		case "attribute":
			if _, err = stm[`CreateCheckConfigurationConstraintAttribute`].Exec(
				conf.Id,
				constr.Attribute.Name,
				constr.Attribute.Value,
			); err != nil {
				break constrloop
			}
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
