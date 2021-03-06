/*-
 * Copyright (c) 2018, Jörg Pernfuß
 * Copyright (c) 2018, 1&1 IONOS SE
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package soma // import "github.com/mjolnir42/soma/internal/soma"

import (
	"database/sql"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/handler"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
)

// JobStatusRead handles read requests for object entities
type JobStatusRead struct {
	Input       chan msg.Request
	Shutdown    chan struct{}
	handlerName string
	conn        *sql.DB
	stmtList    *sql.Stmt
	stmtShow    *sql.Stmt
	stmtSearch  *sql.Stmt
	appLog      *logrus.Logger
	reqLog      *logrus.Logger
	errLog      *logrus.Logger
}

// newJobStatusRead return a new JobStatusRead handler with input
// buffer of length
func newJobStatusRead(length int) (string, *JobStatusRead) {
	r := &JobStatusRead{}
	r.handlerName = generateHandlerName() + `_r`
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return r.handlerName, r
}

// Register initializes resources provided by the Soma app
func (r *JobStatusRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// RegisterRequests links the handler inside the handlermap to the requests
// it processes
func (r *JobStatusRead) RegisterRequests(hmap *handler.Map) {
	for _, action := range []string{
		msg.ActionList,
		msg.ActionShow,
		msg.ActionSearch,
	} {
		hmap.Request(msg.SectionJobStatusMgmt, action, r.handlerName)
	}
}

// Intake exposes the Input channel as part of the handler interface
func (r *JobStatusRead) Intake() chan msg.Request {
	return r.Input
}

// PriorityIntake aliases Intake as part of the handler interface
func (r *JobStatusRead) PriorityIntake() chan msg.Request {
	return r.Intake()
}

// Run is the event loop for JobStatusRead
func (r *JobStatusRead) Run() {
	var err error

	for statement, prepStmt := range map[string]**sql.Stmt{
		stmt.JobStatusMgmtList:   &r.stmtList,
		stmt.JobStatusMgmtShow:   &r.stmtShow,
		stmt.JobStatusMgmtSearch: &r.stmtSearch,
	} {
		if *prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`jobType`, err, stmt.Name(statement))
		}
		defer (*prepStmt).Close()
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
func (r *JobStatusRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	logRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionSearch:
		r.search(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	default:
	}

	q.Reply <- result
}

// list returns all job types
func (r *JobStatusRead) list(q *msg.Request, mr *msg.Result) {
	var (
		err  error
		rows *sql.Rows
		id   string
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&id,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.JobStatus = append(mr.JobStatus, proto.JobStatus{
			ID: id,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// show returns the details of a specific job type
func (r *JobStatusRead) show(q *msg.Request, mr *msg.Result) {
	var id, name, userName string
	var err error
	var ts time.Time

	if err = r.stmtShow.QueryRow(
		q.JobStatus.ID,
	).Scan(
		&id,
		&name,
		&userName,
		&ts,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.JobStatus = append(mr.JobStatus, proto.JobStatus{
		ID:   id,
		Name: name,
		Details: &proto.JobStatusDetails{
			Creation: &proto.DetailsCreation{
				CreatedAt: ts.Format(msg.RFC3339Milli),
				CreatedBy: userName,
			},
		},
	})
	mr.OK()
}

// search returns a job type by ID or Name
func (r *JobStatusRead) search(q *msg.Request, mr *msg.Result) {
	var id, name string
	var err error
	var searchID, searchName sql.NullString

	if q.Search.JobStatus.ID != `` {
		searchID.String = q.Search.JobStatus.ID
		searchID.Valid = true
	}
	if q.Search.JobStatus.Name != `` {
		searchName.String = q.Search.JobStatus.Name
		searchName.Valid = true
	}

	if err = r.stmtSearch.QueryRow(
		&searchID,
		&searchName,
	).Scan(
		&id,
		&name,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.JobStatus = append(mr.JobStatus, proto.JobStatus{
		ID:   id,
		Name: name,
	})
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *JobStatusRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
