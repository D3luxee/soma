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

// LevelRead handles read requests for alert levels
type LevelRead struct {
	Input    chan msg.Request
	Shutdown chan struct{}
	conn     *sql.DB
	stmtList *sql.Stmt
	stmtShow *sql.Stmt
	appLog   *logrus.Logger
	reqLog   *logrus.Logger
	errLog   *logrus.Logger
}

// newLevelRead return a new LevelRead handler with input buffer of
// length
func newLevelRead(length int) (r *LevelRead) {
	r = &LevelRead{}
	r.Input = make(chan msg.Request, length)
	r.Shutdown = make(chan struct{})
	return
}

// Register initializes resources provided by the Soma app
func (r *LevelRead) Register(c *sql.DB, l ...*logrus.Logger) {
	r.conn = c
	r.appLog = l[0]
	r.reqLog = l[1]
	r.errLog = l[2]
}

// Run is the event loop for LevelRead
func (r *LevelRead) Run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.LevelList: r.stmtList,
		stmt.LevelShow: r.stmtShow,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`level`, err, stmt.Name(statement))
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
func (r *LevelRead) process(q *msg.Request) {
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

// list returns all alert levels
func (r *LevelRead) list(q *msg.Request, mr *msg.Result) {
	var (
		level, short string
		rows         *sql.Rows
		err          error
	)

	if rows, err = r.stmtList.Query(); err != nil {
		mr.ServerError(err, q.Section)
		return
	}

	for rows.Next() {
		if err = rows.Scan(&level, &short); err != nil {
			rows.Close()
			mr.ServerError(err, q.Section)
			return
		}
		mr.Level = append(mr.Level, proto.Level{
			Name:      level,
			ShortName: short,
		})
	}
	if err = rows.Err(); err != nil {
		mr.ServerError(err)
		return
	}
	mr.OK()
}

// show returns the details of a specific alert levels
func (r *LevelRead) show(q *msg.Request, mr *msg.Result) {
	var (
		level, short string
		numeric      uint16
		err          error
	)

	if err = r.stmtShow.QueryRow(
		q.Level.Name,
	).Scan(
		&level,
		&short,
		&numeric,
	); err == sql.ErrNoRows {
		mr.NotFound(err, q.Section)
		return
	} else if err != nil {
		mr.ServerError(err, q.Section)
		return
	}
	mr.Level = append(mr.Level, proto.Level{
		Name:      level,
		ShortName: short,
		Numeric:   numeric,
	})
	mr.OK()
}

// ShutdownNow signals the handler to shut down
func (r *LevelRead) ShutdownNow() {
	close(r.Shutdown)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
