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

// NodeRead handles read requests for nodes
type NodeRead struct {
	Input           chan msg.Request
	Shutdown        chan struct{}
	conn            *sql.DB
	stmtList        *sql.Stmt
	stmtShow        *sql.Stmt
	stmtShowConfig  *sql.Stmt
	stmtSync        *sql.Stmt
	stmtPropOncall  *sql.Stmt
	stmtPropService *sql.Stmt
	stmtPropSystem  *sql.Stmt
	stmtPropCustom  *sql.Stmt
	appLog          *logrus.Logger
	reqLog          *logrus.Logger
	errLog          *logrus.Logger
}

// newNodeRead return a new NodeRead handler with input buffer of length
func newNodeRead(length int) (r *NodeRead) {
	r = &NodeRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// Register initializes resources provided by the Soma app
func (r *NodeRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// Run is the event loop for NodeRead
func (r *NodeRead) Run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.NodeList:       r.stmtList,
		stmt.NodeShow:       r.stmtShow,
		stmt.NodeShowConfig: r.stmtShowConfig,
		stmt.NodeSync:       r.stmtSync,
		stmt.NodeOncProps:   r.stmtPropOncall,
		stmt.NodeSvcProps:   r.stmtPropService,
		stmt.NodeSysProps:   r.stmtPropSystem,
		stmt.NodeCstProps:   r.stmtPropCustom,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`node`, err, stmt.Name(statement))
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
func (r *NodeRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionSync:
		r.sync(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionShowConfig:
		r.showConfig(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all nodes
func (r *NodeRead) list(q *msg.Request, mr *msg.Result) {
	var (
		rows             *sql.Rows
		err              error
		nodeID, nodeName string
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&nodeID, &nodeName); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Node = append(mr.Node, proto.Node{
			Id:   nodeID,
			Name: nodeName,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// sync returns all nodes with all details attached
func (r *NodeRead) sync(q *msg.Request, mr *msg.Result) {
	var (
		rows                                   *sql.Rows
		err                                    error
		nodeID, nodeName, nodeTeam, nodeServer string
		nodeAsset                              int
		nodeOnline, nodeDeleted                bool
	)

	if rows, err = r.stmtSync.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&nodeID,
			&nodeAsset,
			&nodeName,
			&nodeTeam,
			&nodeServer,
			&nodeOnline,
			&nodeDeleted,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Node = append(mr.Node, proto.Node{
			Id:        nodeID,
			AssetId:   uint64(nodeAsset),
			Name:      nodeName,
			TeamId:    nodeTeam,
			ServerId:  nodeServer,
			IsOnline:  nodeOnline,
			IsDeleted: nodeDeleted,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns the details for a specific node
func (r *NodeRead) show(q *msg.Request, mr *msg.Result) {
	var (
		err                                    error
		nodeID, nodeName, nodeTeam, nodeServer string
		repositoryID, bucketID, nodeState      string
		nodeOnline, nodeDeleted                bool
		nodeAsset                              int
		node                                   proto.Node
		tx                                     *sql.Tx
		checkConfigs                           *[]proto.CheckConfig
	)

	if err = r.stmtShow.QueryRow(
		q.Node.Id,
	).Scan(
		&nodeID,
		&nodeAsset,
		&nodeName,
		&nodeTeam,
		&nodeServer,
		&nodeState,
		&nodeOnline,
		&nodeDeleted,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		goto fail
	}
	node = proto.Node{
		Id:        nodeID,
		AssetId:   uint64(nodeAsset),
		Name:      nodeName,
		TeamId:    nodeTeam,
		ServerId:  nodeServer,
		State:     nodeState,
		IsOnline:  nodeOnline,
		IsDeleted: nodeDeleted,
	}

	// add configuration data
	if err = r.stmtShowConfig.QueryRow(
		q.Node.Id,
	).Scan(
		&nodeID,
		&nodeName,
		&bucketID,
		&repositoryID,
	); err == sql.ErrNoRows {
		// sql.ErrNoRows means the node is unassigned, which is
		// valid and not an error. But an unconfigured node can
		// not have properties or checks, which means the request
		// is done.
		mr.OK()
		return
	} else if err != nil {
		goto fail
	}
	// node is assigned in this codepath
	node.Config = &proto.NodeConfig{
		RepositoryId: repositoryID,
		BucketId:     bucketID,
	}

	// fetch node properties
	node.Properties = &[]proto.Property{}

	if err = r.oncallProperties(&node); err != nil {
		goto fail
	}
	if err = r.serviceProperties(&node); err != nil {
		goto fail
	}
	if err = r.systemProperties(&node); err != nil {
		goto fail
	}
	if err = r.customProperties(&node); err != nil {
		goto fail
	}
	if len(*node.Properties) == 0 {
		// trigger ,omitempty in JSON export
		node.Properties = nil
	}

	// add check configuration and instance information
	if tx, err = r.conn.Begin(); err != nil {
		goto fail
	}
	if checkConfigs, err = exportCheckConfigObjectTX(
		tx,
		q.Node.Id,
	); err != nil {
		tx.Rollback()
		goto fail
	}
	if checkConfigs != nil && len(*checkConfigs) > 0 {
		node.Details = &proto.Details{
			CheckConfigs: checkConfigs,
		}
	}

	mr.Node = append(mr.Node, node)
	mr.OK()
	return

fail:
	mr.ServerError(err, q.Section)
}

// showConfig returns the repository configuration of the node
func (r *NodeRead) showConfig(q *msg.Request, mr *msg.Result) {
	var (
		err                                      error
		nodeID, nodeName, repositoryID, bucketID string
	)
	if err = r.stmtShowConfig.QueryRow(
		q.Node.Id,
	).Scan(
		&nodeID,
		&nodeName,
		&bucketID,
		&repositoryID,
	); err == sql.ErrNoRows {
		// TODO need a better way to transport 'unassigned'
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Node = append(mr.Node, proto.Node{
		Id:   nodeID,
		Name: nodeName,
		Config: &proto.NodeConfig{
			RepositoryId: repositoryID,
			BucketId:     bucketID,
		},
	})
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *NodeRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
