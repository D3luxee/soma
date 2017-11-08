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
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/Sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
)

// JobRead handles read requests for jobs
type JobRead struct {
	Input                     chan msg.Request
	Shutdown                  chan struct{}
	conn                      *sql.DB
	stmtListAllOutstanding    *sql.Stmt
	stmtListScopedOutstanding *sql.Stmt
	stmtResultByID            *sql.Stmt
	stmtResultByIDList        *sql.Stmt
	appLog                    *logrus.Logger
	reqLog                    *logrus.Logger
	errLog                    *logrus.Logger
}

// newJobRead return a new JobRead handler with input buffer of
// length
func newJobRead(length int) (r *JobRead) {
	r = &JobRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// Register initializes resources provided by the Soma app
func (r *JobRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// Intake exposes the Input channel as part of the handler interface
func (r *JobRead) Intake() chan msg.Request {
	return r.Input
}

// Run is the event loop for JobRead
func (r *JobRead) Run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.ListAllOutstandingJobs:    r.stmtListAllOutstanding,
		stmt.ListScopedOutstandingJobs: r.stmtListScopedOutstanding,
		stmt.JobResultForId:            r.stmtResultByID,
		stmt.JobResultsForList:         r.stmtResultByIDList,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`jobs`, err, stmt.Name(statement))
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
func (r *JobRead) process(q *msg.Request) {
	result := msg.FromRequest(q)
	msgRequest(r.reqLog, q)

	switch q.Action {
	case msg.ActionList:
		r.list(q, &result)
	case msg.ActionAll:
		r.all(q, &result)
	case msg.ActionShow:
		r.show(q, &result)
	case msg.ActionSearchByList:
		r.search(q, &result)
	default:
		result.UnknownRequest(q)
	}

	q.Reply <- result
}

// list the outstanding jobs for a specific user
func (r *JobRead) list(q *msg.Request, mr *msg.Result) {
	var (
		rows           *sql.Rows
		err            error
		jobID, jobType string
	)

	if rows, err = r.stmtListScopedOutstanding.Query(
		q.AuthUser,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&jobID,
			&jobType,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Job = append(mr.Job,
			proto.Job{
				Id:   jobID,
				Type: jobType,
			},
		)
	}
	if rows.Err() != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// all returns a list of all outstanding jobs
func (r *JobRead) all(q *msg.Request, mr *msg.Result) {
	var (
		rows           *sql.Rows
		err            error
		jobID, jobType string
	)

	// section: runtime
	if rows, err = r.stmtListAllOutstanding.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&jobID,
			&jobType,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Job = append(mr.Job,
			proto.Job{
				Id:   jobID,
				Type: jobType,
			},
		)
	}
	if rows.Err() != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// show returns the details about a specific job
func (r *JobRead) show(q *msg.Request, mr *msg.Result) {
	var (
		err                                                error
		jobID, jobType, jobStatus, jobResult, repositoryID string
		jobError, jobSpec, teamID, userID                  string
		jobSerial                                          int
		jobQueued                                          time.Time
		jobStarted, jobFinished                            pq.NullTime
	)

	if err = r.stmtResultByID.QueryRow(
		q.Job.Id,
	).Scan(
		&jobID,
		&jobStatus,
		&jobResult,
		&jobType,
		&jobSerial,
		&repositoryID,
		&userID,
		&teamID,
		&jobQueued,
		&jobStarted,
		&jobFinished,
		&jobError,
		&jobSpec,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	job := proto.Job{
		Id:           jobID,
		Status:       jobStatus,
		Result:       jobResult,
		Type:         jobType,
		Serial:       jobSerial,
		RepositoryId: repositoryID,
		UserId:       userID,
		TeamId:       teamID,
		Error:        jobError,
	}
	job.TsQueued = jobQueued.Format(msg.RFC3339Milli)
	if jobStarted.Valid {
		job.TsStarted = jobStarted.Time.Format(msg.RFC3339Milli)
	}
	if jobFinished.Valid {
		job.TsFinished = jobFinished.Time.Format(msg.RFC3339Milli)
	}
	if q.Flag.JobDetail {
		job.Details = &proto.JobDetails{
			Specification: jobSpec,
		}
	}
	mr.Job = []proto.Job{job}
	mr.OK()
}

// search returns the details for a list of jobs
func (r *JobRead) search(q *msg.Request, mr *msg.Result) {
	var (
		rows                                               *sql.Rows
		err                                                error
		jobID, jobType, jobStatus, jobResult, repositoryID string
		userID, teamID, jobError, jobSpec, idList          string
		jobSerial                                          int
		jobQueued                                          time.Time
		jobStarted, jobFinished                            pq.NullTime
	)

	idList = fmt.Sprintf("{%s}", strings.Join(q.Search.Job.IdList, `,`))
	if rows, err = r.stmtResultByIDList.Query(
		idList,
	); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(
			&jobID,
			&jobStatus,
			&jobResult,
			&jobType,
			&jobSerial,
			&repositoryID,
			&userID,
			&teamID,
			&jobQueued,
			&jobStarted,
			&jobFinished,
			&jobError,
			&jobSpec,
		); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		job := proto.Job{
			Id:           jobID,
			Status:       jobStatus,
			Result:       jobResult,
			Type:         jobType,
			Serial:       jobSerial,
			RepositoryId: repositoryID,
			UserId:       userID,
			TeamId:       teamID,
			Error:        jobError,
		}
		job.TsQueued = jobQueued.Format(msg.RFC3339Milli)
		if jobStarted.Valid {
			job.TsStarted = jobStarted.Time.Format(msg.RFC3339Milli)
		}
		if jobFinished.Valid {
			job.TsFinished = jobFinished.Time.Format(msg.RFC3339Milli)
		}
		if q.Flag.JobDetail && q.Search.IsDetailed {
			job.Details = &proto.JobDetails{
				Specification: jobSpec,
			}
		}
		mr.Job = append(mr.Job, job)
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *JobRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
