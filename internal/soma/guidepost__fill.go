/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package soma

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
	"github.com/satori/go.uuid"
)

//
func (g *GuidePost) fillReqData(q *msg.Request) (error, bool) {
	switch {
	case strings.Contains(q.Action, "add_service_property_to_"):
		return g.fillServiceAttributes(q)
	case q.Action == `assign_node`:
		return g.fillNode(q)
	case q.Action == `remove_check`:
		return g.fillCheckDeleteInfo(q)
	case strings.HasPrefix(q.Action, `delete_`) &&
		strings.Contains(q.Action, `_property_from_`):
		return g.fillPropertyDeleteInfo(q)
	case strings.HasPrefix(q.Action, `add_check_to_`):
		return g.fillCheckConfigId(q)
	default:
		return nil, false
	}
}

// generate CheckConfigId
func (g *GuidePost) fillCheckConfigId(q *msg.Request) (error, bool) {
	q.CheckConfig.Id = uuid.NewV4().String()
	return nil, false
}

// Populate the node structure with data, overwriting the client
// submitted values.
func (g *GuidePost) fillNode(q *msg.Request) (error, bool) {
	var (
		err                      error
		ndName, ndTeam, ndServer string
		ndAsset                  int64
		ndOnline, ndDeleted      bool
	)
	if err = g.stmtNodeDetails.QueryRow(q.Node.Id).Scan(
		&ndAsset,
		&ndName,
		&ndTeam,
		&ndServer,
		&ndOnline,
		&ndDeleted,
	); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("Node not found: %s", q.Node.Id), true
		}
		return err, false
	}
	q.Node.AssetId = uint64(ndAsset)
	q.Node.Name = ndName
	q.Node.TeamId = ndTeam
	q.Node.ServerId = ndServer
	q.Node.IsOnline = ndOnline
	q.Node.IsDeleted = ndDeleted
	return nil, false
}

// load authoritative copy of the service attributes from the
// database. Replaces whatever the client sent in.
func (g *GuidePost) fillServiceAttributes(q *msg.Request) (error, bool) {
	var (
		service, attr, val, svName, svTeam, repoID string
		rows                                       *sql.Rows
		err                                        error
		nf                                         bool
	)
	attrs := []proto.ServiceAttribute{}

	switch q.Section {
	case msg.SectionRepository:
		svName = (*q.Repository.Properties)[0].Service.Name
		svTeam = (*q.Repository.Properties)[0].Service.TeamId
	case msg.SectionBucket:
		svName = (*q.Bucket.Properties)[0].Service.Name
		svTeam = (*q.Bucket.Properties)[0].Service.TeamId
	case msg.SectionGroup:
		svName = (*q.Group.Properties)[0].Service.Name
		svTeam = (*q.Group.Properties)[0].Service.TeamId
	case msg.SectionCluster:
		svName = (*q.Cluster.Properties)[0].Service.Name
		svTeam = (*q.Cluster.Properties)[0].Service.TeamId
	case msg.SectionNode:
		svName = (*q.Node.Properties)[0].Service.Name
		svTeam = (*q.Node.Properties)[0].Service.TeamId
	}

	// ignore error since it would have been caught by GuidePost
	repoID, _, _, _ = g.extractRouting(q)

	// validate the tuple (repo, team, service) is valid
	if err = g.stmtServiceLookup.QueryRow(
		repoID, svName, svTeam,
	).Scan(
		&service,
	); err != nil {
		if err == sql.ErrNoRows {
			nf = true
			err = fmt.Errorf("Requested service %s not available for team %s",
				svName, svTeam)
		}
		goto abort
	}

	// load attributes
	if rows, err = g.stmtServiceAttributes.Query(
		repoID, svName, svTeam,
	); err != nil {
		goto abort
	}
	defer rows.Close()

attrloop:
	for rows.Next() {
		if err = rows.Scan(&attr, &val); err != nil {
			break attrloop
		}
		attrs = append(attrs, proto.ServiceAttribute{
			Name:  attr,
			Value: val,
		})
	}
abort:
	if err != nil {
		return err, nf
	}
	// not aborted: set the loaded attributes
	switch q.Section {
	case msg.SectionRepository:
		(*q.Repository.Properties)[0].Service.Attributes = attrs
	case msg.SectionBucket:
		(*q.Bucket.Properties)[0].Service.Attributes = attrs
	case msg.SectionGroup:
		(*q.Group.Properties)[0].Service.Attributes = attrs
	case msg.SectionCluster:
		(*q.Cluster.Properties)[0].Service.Attributes = attrs
	case msg.SectionNode:
		(*q.Node.Properties)[0].Service.Attributes = attrs
	}
	return nil, false
}

// if the request is a check deletion, populate required IDs
func (g *GuidePost) fillCheckDeleteInfo(q *msg.Request) (error, bool) {
	var delObjId, delObjTyp, delSrcChkId string
	var err error

	if err = g.stmtCheckDetailsForDelete.QueryRow(
		q.CheckConfig.Id,
		q.CheckConfig.RepositoryId,
	).Scan(
		&delObjId,
		&delObjTyp,
		&delSrcChkId,
	); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf(
				"Failed to find source check for config %s",
				q.CheckConfig.Id), true
		}
		return err, false
	}
	q.CheckConfig.ObjectId = delObjId
	q.CheckConfig.ObjectType = delObjTyp
	q.CheckConfig.ExternalId = delSrcChkId
	q.Action = fmt.Sprintf("remove_check_from_%s", delObjTyp)
	return nil, false
}

// if the request is a property deletion, populate required IDs
func (g *GuidePost) fillPropertyDeleteInfo(q *msg.Request) (error, bool) {
	var (
		err                                             error
		row                                             *sql.Row
		queryStmt, view, sysProp, value, cstId, cstProp string
		svcProp, oncId, oncName                         string
		oncNumber                                       int
	)

	// select SQL statement
	switch q.Action {
	case `delete_system_property_from_repository`:
		queryStmt = stmt.RepoSystemPropertyForDelete
	case `delete_custom_property_from_repository`:
		queryStmt = stmt.RepoCustomPropertyForDelete
	case `delete_service_property_from_repository`:
		queryStmt = stmt.RepoServicePropertyForDelete
	case `delete_oncall_property_from_repository`:
		queryStmt = stmt.RepoOncallPropertyForDelete
	case `delete_system_property_from_bucket`:
		queryStmt = stmt.BucketSystemPropertyForDelete
	case `delete_custom_property_from_bucket`:
		queryStmt = stmt.BucketCustomPropertyForDelete
	case `delete_service_property_from_bucket`:
		queryStmt = stmt.BucketServicePropertyForDelete
	case `delete_oncall_property_from_bucket`:
		queryStmt = stmt.BucketOncallPropertyForDelete
	case `delete_system_property_from_group`:
		queryStmt = stmt.GroupSystemPropertyForDelete
	case `delete_custom_property_from_group`:
		queryStmt = stmt.GroupCustomPropertyForDelete
	case `delete_service_property_from_group`:
		queryStmt = stmt.GroupServicePropertyForDelete
	case `delete_oncall_property_from_group`:
		queryStmt = stmt.GroupOncallPropertyForDelete
	case `delete_system_property_from_cluster`:
		queryStmt = stmt.ClusterSystemPropertyForDelete
	case `delete_custom_property_from_cluster`:
		queryStmt = stmt.ClusterCustomPropertyForDelete
	case `delete_service_property_from_cluster`:
		queryStmt = stmt.ClusterServicePropertyForDelete
	case `delete_oncall_property_from_cluster`:
		queryStmt = stmt.ClusterOncallPropertyForDelete
	case `delete_system_property_from_node`:
		queryStmt = stmt.NodeSystemPropertyForDelete
	case `delete_custom_property_from_node`:
		queryStmt = stmt.NodeCustomPropertyForDelete
	case `delete_service_property_from_node`:
		queryStmt = stmt.NodeServicePropertyForDelete
	case `delete_oncall_property_from_node`:
		queryStmt = stmt.NodeOncallPropertyForDelete
	}

	// execute and scan
	switch q.Section {
	case msg.SectionRepository:
		row = g.conn.QueryRow(queryStmt,
			(*q.Repository.Properties)[0].SourceInstanceId)
	case msg.SectionBucket:
		row = g.conn.QueryRow(queryStmt,
			(*q.Bucket.Properties)[0].SourceInstanceId)
	case msg.SectionGroup:
		row = g.conn.QueryRow(queryStmt,
			(*q.Group.Properties)[0].SourceInstanceId)
	case msg.SectionCluster:
		row = g.conn.QueryRow(queryStmt,
			(*q.Cluster.Properties)[0].SourceInstanceId)
	case msg.SectionNode:
		row = g.conn.QueryRow(queryStmt,
			(*q.Node.Properties)[0].SourceInstanceId)
	}
	switch {
	case strings.HasPrefix(q.Action, `delete_system_`):
		err = row.Scan(&view, &sysProp, &value)

	case strings.HasPrefix(q.Action, `delete_custom_`):
		err = row.Scan(&view, &cstId, &value, &cstProp)

	case strings.HasPrefix(q.Action, `delete_service_`):
		err = row.Scan(&view, &svcProp)

	case strings.HasPrefix(q.Action, `delete_oncall_`):
		err = row.Scan(&view, &oncId, &oncName, &oncNumber)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf(
				"Failed to find source property for %s",
				(*q.Repository.Properties)[0].SourceInstanceId), true
		}
		return err, false
	}

	// assemble and set results: property specification
	var (
		pSys *proto.PropertySystem
		pCst *proto.PropertyCustom
		pSvc *proto.PropertyService
		pOnc *proto.PropertyOncall
	)
	switch {
	case strings.HasPrefix(q.Action, `delete_system_`):
		pSys = &proto.PropertySystem{
			Name:  sysProp,
			Value: value,
		}
	case strings.HasPrefix(q.Action, `delete_custom_`):
		pCst = &proto.PropertyCustom{
			Id:    cstId,
			Name:  cstProp,
			Value: value,
		}
	case strings.HasPrefix(q.Action, `delete_service_`):
		pSvc = &proto.PropertyService{
			Name: svcProp,
		}
	case strings.HasPrefix(q.Action, `delete_oncall_`):
		num := strconv.Itoa(oncNumber)
		pOnc = &proto.PropertyOncall{
			Id:     oncId,
			Name:   oncName,
			Number: num,
		}
	}

	// assemble and set results: view
	switch {
	case strings.HasSuffix(q.Action, `_repository`):
		(*q.Repository.Properties)[0].View = view
	case strings.HasSuffix(q.Action, `_bucket`):
		(*q.Bucket.Properties)[0].View = view
	case strings.HasSuffix(q.Action, `_group`):
		(*q.Group.Properties)[0].View = view
	case strings.HasSuffix(q.Action, `_cluster`):
		(*q.Cluster.Properties)[0].View = view
	case strings.HasSuffix(q.Action, `_node`):
		(*q.Node.Properties)[0].View = view
	}

	// final assembly step
	switch q.Action {
	case `delete_system_property_from_repository`:
		(*q.Repository.Properties)[0].System = pSys
	case `delete_custom_property_from_repository`:
		(*q.Repository.Properties)[0].Custom = pCst
	case `delete_service_property_from_repository`:
		(*q.Repository.Properties)[0].Service = pSvc
	case `delete_oncall_property_from_repository`:
		(*q.Repository.Properties)[0].Oncall = pOnc

	case `delete_system_property_from_bucket`:
		(*q.Bucket.Properties)[0].System = pSys
	case `delete_custom_property_from_bucket`:
		(*q.Bucket.Properties)[0].Custom = pCst
	case `delete_service_property_from_bucket`:
		(*q.Bucket.Properties)[0].Service = pSvc
	case `delete_oncall_property_from_bucket`:
		(*q.Bucket.Properties)[0].Oncall = pOnc

	case `delete_system_property_from_group`:
		(*q.Group.Properties)[0].System = pSys
	case `delete_custom_property_from_group`:
		(*q.Group.Properties)[0].Custom = pCst
	case `delete_service_property_from_group`:
		(*q.Group.Properties)[0].Service = pSvc
	case `delete_oncall_property_from_group`:
		(*q.Group.Properties)[0].Oncall = pOnc

	case `delete_system_property_from_cluster`:
		(*q.Cluster.Properties)[0].System = pSys
	case `delete_custom_property_from_cluster`:
		(*q.Cluster.Properties)[0].Custom = pCst
	case `delete_service_property_from_cluster`:
		(*q.Cluster.Properties)[0].Service = pSvc
	case `delete_oncall_property_from_cluster`:
		(*q.Cluster.Properties)[0].Oncall = pOnc

	case `delete_system_property_from_node`:
		(*q.Node.Properties)[0].System = pSys
	case `delete_custom_property_from_node`:
		(*q.Node.Properties)[0].Custom = pCst
	case `delete_service_property_from_node`:
		(*q.Node.Properties)[0].Service = pSvc
	case `delete_oncall_property_from_node`:
		(*q.Node.Properties)[0].Oncall = pOnc
	}
	return nil, false
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix