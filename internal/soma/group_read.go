/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 * Copyright (c) 2016, 1&1 Internet SE
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package soma

import (
	"database/sql"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
)

// GroupRead handles read requests for groups
type GroupRead struct {
	Input                 chan msg.Request
	Shutdown              chan struct{}
	conn                  *sql.DB
	stmtList              *sql.Stmt
	stmtShow              *sql.Stmt
	stmtMemberListGroup   *sql.Stmt
	stmtMemberListCluster *sql.Stmt
	stmtMemberListNode    *sql.Stmt
	stmtPropOncall        *sql.Stmt
	stmtPropService       *sql.Stmt
	stmtPropSystem        *sql.Stmt
	stmtPropCustom        *sql.Stmt
	appLog                *logrus.Logger
	reqLog                *logrus.Logger
	errLog                *logrus.Logger
}

// newGroupRead returns a new GroupRead handler with input
// buffer of length
func newGroupRead(length int) (r *GroupRead) {
	r = &GroupRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// register initializes resources provided by the Soma app
func (r *GroupRead) register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// run is the event loop for GroupRead
func (r *GroupRead) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.GroupList:              r.stmtList,
		stmt.GroupShow:              r.stmtShow,
		stmt.GroupMemberGroupList:   r.stmtMemberListGroup,
		stmt.GroupMemberClusterList: r.stmtMemberListCluster,
		stmt.GroupMemberNodeList:    r.stmtMemberListNode,
		stmt.GroupOncProps:          r.stmtPropOncall,
		stmt.GroupSvcProps:          r.stmtPropService,
		stmt.GroupSysProps:          r.stmtPropSystem,
		stmt.GroupCstProps:          r.stmtPropCustom,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`group`, err, stmt.Name(statement))
		}
		defer prepStmt.Close()
	}

runloop:
	for {
		select {
		case <-r.Shutdown:
			break runloop
		case req := <-r.Input:
			go func() {
				r.process(&req)
			}()
		}
	}
}

// process is the request dispatcher
func (r *GroupRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionMemberList:
		r.memberList(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all groups
func (r *GroupRead) list(q *msg.Request, mr *msg.Result) {
	var (
		groupID, groupName, bucketID string
		rows                         *sql.Rows
		err                          error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&groupID,
			&groupName,
			&bucketID,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Group = append(mr.Group, proto.Group{
			Id:       groupID,
			Name:     groupName,
			BucketId: bucketID,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns the details of a specific group
func (r *GroupRead) show(q *msg.Request, mr *msg.Result) {
	var (
		groupID, groupName, groupState string
		bucketID, teamID               string
		err                            error
		tx                             *sql.Tx
		checkConfigs                   *[]proto.CheckConfig
		group                          proto.Group
	)

	if err = r.stmtShow.QueryRow(
		q.Group.Id,
	).Scan(
		&groupID,
		&bucketID,
		&groupName,
		&groupState,
		&teamID,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		goto fail
	}
	group = proto.Group{
		Id:          groupID,
		Name:        groupName,
		BucketId:    bucketID,
		ObjectState: groupState,
		TeamId:      teamID,
	}

	// add properties
	group.Properties = &[]proto.Property{}

	if err = r.oncallProperties(&group); err != nil {
		goto fail
	}
	if err = r.serviceProperties(&group); err != nil {
		goto fail
	}
	if err = r.systemProperties(&group); err != nil {
		goto fail
	}
	if err = r.customProperties(&group); err != nil {
		goto fail
	}
	if len(*group.Properties) == 0 {
		// trigger ,omitempty in JSON export
		group.Properties = nil
	}

	// add check configuration and instance information
	if tx, err = r.conn.Begin(); err != nil {
		goto fail
	}
	if checkConfigs, err = exportCheckConfigObjectTX(
		tx,
		q.Group.Id,
	); err != nil {
		tx.Rollback()
		goto fail
	}
	if checkConfigs != nil && len(*checkConfigs) > 0 {
		group.Details = &proto.Details{
			CheckConfigs: checkConfigs,
		}
	}

	mr.Group = append(mr.Group, group)
	mr.OK()
	return

fail:
	mr.ServerError(err, q.Section)

}

// memberList returns the group members
func (r *GroupRead) memberList(q *msg.Request, mr *msg.Result) {
	var (
		group                              proto.Group
		groupName                          string
		memberGroupID, memberGroupName     string
		memberClusterID, memberClusterName string
		memberNodeID, memberNodeName       string
		err                                error
		rows                               *sql.Rows
	)

	group.Id = q.Group.Id
	group.MemberGroups = &[]proto.Group{}
	group.MemberClusters = &[]proto.Cluster{}
	group.MemberNodes = &[]proto.Node{}

	// fetch member groups
	if rows, err = r.stmtMemberListGroup.Query(
		q.Group.Id,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&memberGroupID,
			&memberGroupName,
			&groupName,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		group.Name = groupName
		*group.MemberGroups = append(*group.MemberGroups, proto.Group{
			Id:   memberGroupID,
			Name: memberGroupName,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if len(*group.MemberGroups) == 0 {
		// trigger ,omitempty in JSON export
		group.MemberGroups = nil
	}

	// fetch member clusters
	if rows, err = r.stmtMemberListCluster.Query(
		q.Group.Id,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&memberClusterID,
			&memberClusterName,
			&groupName,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		*group.MemberClusters = append(*group.MemberClusters,
			proto.Cluster{
				Id:   memberClusterID,
				Name: memberClusterName,
			})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if len(*group.MemberClusters) == 0 {
		// trigger ,omitempty in JSON export
		group.MemberClusters = nil
	}

	// fetch member nodes
	if rows, err = r.stmtMemberListNode.Query(
		q.Group.Id,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&memberNodeID,
			&memberNodeName,
			&groupName,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		*group.MemberNodes = append(*group.MemberNodes,
			proto.Node{
				Id:   memberNodeID,
				Name: memberNodeName,
			})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if len(*group.MemberNodes) == 0 {
		// trigger ,omitempty in JSON export
		group.MemberNodes = nil
	}
	mr.Group = append(mr.Group, group)
	mr.OK()
}

// shutdownNow signals the handler to shut down
func (r *GroupRead) shutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix