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
	uuid "github.com/satori/go.uuid"
)

// MonitoringWrite handles write requests for monitoring systems
type MonitoringWrite struct {
	Input      chan msg.Request
	Shutdown   chan struct{}
	conn       *sql.DB
	stmtCreate *sql.Stmt
	stmtDelete *sql.Stmt
	appLog     *logrus.Logger
	reqLog     *logrus.Logger
	errLog     *logrus.Logger
}

// newMonitoringWrite return a new MonitoringWrite handler with
// input buffer of length
func newMonitoringWrite(length int) (w *MonitoringWrite) {
	w = &MonitoringWrite{}
	w.Input = make(chan msg.Request, length)
	w.Shutdown = make(chan struct{})
	return
}

// register initializes resources provided by the Soma app
func (w *MonitoringWrite) register(c *sql.DB, l ...*logrus.Logger) {
	w.conn = c
	w.appLog = l[0]
	w.reqLog = l[1]
	w.errLog = l[2]
}

// run is the event loop for MonitoringWrite
func (w *MonitoringWrite) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.MonitoringSystemAdd:    w.stmtCreate,
		stmt.MonitoringSystemRemove: w.stmtDelete,
	} {
		if prepStmt, err = w.conn.Prepare(statement); err != nil {
			w.errLog.Fatal(`monitoring`, err, stmt.Name(statement))
		}
		defer prepStmt.Close()
	}

runloop:
	for {
		select {
		case <-w.Shutdown:
			break runloop
		case req := <-w.Input:
			w.process(&req)
		}
	}
}

// process is the request dispatcher
func (w *MonitoringWrite) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(w.reqLog, q)

	switch q.Action {
	case msg.ActionCreate:
		w.create(q, &result)
	case msg.ActionDelete:
		w.delete(q, &result)
	default:
		result.UnknownRequest(q)
	}

	q.Reply <- result
}

// create adds a new monitoring system
func (w *MonitoringWrite) create(q *msg.Request, mr *msg.Result) {
	var (
		err      error
		res      sql.Result
		callback sql.NullString
	)

	q.Monitoring.Id = uuid.NewV4().String()
	if q.Monitoring.Callback != `` {
		callback = sql.NullString{
			String: q.Monitoring.Callback,
			Valid:  true,
		}
	}
	if res, err = w.stmtCreate.Exec(
		q.Monitoring.Id,
		q.Monitoring.Name,
		q.Monitoring.Mode,
		q.Monitoring.Contact,
		q.Monitoring.TeamId,
		callback,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if mr.RowCnt(res.RowsAffected()) {
		mr.Monitoring = append(mr.Monitoring, q.Monitoring)
	}
}

// delete removes a monitoring system
func (w *MonitoringWrite) delete(q *msg.Request, mr *msg.Result) {
	var (
		err error
		res sql.Result
	)

	if res, err = w.stmtDelete.Exec(
		q.Monitoring.Id,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	if mr.RowCnt(res.RowsAffected()) {
		mr.Monitoring = append(mr.Monitoring, q.Monitoring)
	}
}

// shutdownNow signals the handler to shut down
func (w *MonitoringWrite) shutdownNow() {
	close(w.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix