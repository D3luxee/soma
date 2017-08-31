/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016-2017, Jörg Pernfuß
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

// AttributeRead handles read requests for attributes
type AttributeRead struct {
	Input    chan msg.Request
	Shutdown chan struct{}
	conn     *sql.DB
	stmtList *sql.Stmt
	stmtShow *sql.Stmt
	appLog   *logrus.Logger
	reqLog   *logrus.Logger
	errLog   *logrus.Logger
}

// newAttributeRead return a new AttributeRead handler with input
// buffer of length
func newAttributeRead(length int) (r *AttributeRead) {
	r = &AttributeRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// register initializes resources provided by the Soma app
func (r *AttributeRead) register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// run is the event loop for AttributeRead
func (r *AttributeRead) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.AttributeList: r.stmtList,
		stmt.AttributeShow: r.stmtShow,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`attribute`, err, stmt.Name(statement))
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
func (r *AttributeRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	default:
		result.UnknownRequest(q)
	}
	q.Reply <- result
}

// list returns all attributes
func (r *AttributeRead) list(q *msg.Request, mr *msg.Result) {
	var (
		err                    error
		attribute, cardinality string
		rows                   *sql.Rows
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&attribute,
			&cardinality,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Attribute = append(mr.Attribute, proto.Attribute{
			Name:        attribute,
			Cardinality: cardinality,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// show returns the details of a specific attribute
func (r *AttributeRead) show(q *msg.Request, mr *msg.Result) {
	var (
		err                    error
		attribute, cardinality string
	)

	if err = r.stmtShow.QueryRow(
		q.Attribute.Name,
	).Scan(
		&attribute,
		&cardinality,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Attribute = append(mr.Attribute, proto.Attribute{
		Name:        attribute,
		Cardinality: cardinality,
	})
	mr.OK()
}

// shutdownNow signals the handler to shut down
func (r *AttributeRead) shutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix