package main

import (
	"database/sql"
	"errors"

	"github.com/1and1/soma/internal/stmt"
	log "github.com/Sirupsen/logrus"
)

// Message structs
type somaObjectStateRequest struct {
	action string
	state  string
	rename string
	reply  chan []somaObjectStateResult
}

type somaObjectStateResult struct {
	err   error
	state string
}

/*  Read Access
 *
 */
type somaObjectStateReadHandler struct {
	input     chan somaObjectStateRequest
	shutdown  chan bool
	conn      *sql.DB
	list_stmt *sql.Stmt
	show_stmt *sql.Stmt
	appLog    *log.Logger
	reqLog    *log.Logger
	errLog    *log.Logger
}

func (r *somaObjectStateReadHandler) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.ObjectStateList: r.list_stmt,
		stmt.ObjectStateShow: r.show_stmt,
	} {
		if prepStmt, err = r.conn.Prepare(statement); err != nil {
			r.errLog.Fatal(`object_state`, err, stmt.Name(statement))
		}
		defer prepStmt.Close()
	}

	for {
		select {
		case <-r.shutdown:
			break
		case req := <-r.input:
			go func() {
				r.process(&req)
			}()
		}
	}
}

func (r *somaObjectStateReadHandler) process(q *somaObjectStateRequest) {
	var state string
	var rows *sql.Rows
	var err error
	result := make([]somaObjectStateResult, 0)

	switch q.action {
	case "list":
		rows, err = r.list_stmt.Query()
		defer rows.Close()
		if err != nil {
			result = append(result, somaObjectStateResult{
				err:   err,
				state: q.state,
			})
			q.reply <- result
			return
		}

		for rows.Next() {
			err = rows.Scan(&state)
			if err != nil {
				result = append(result, somaObjectStateResult{
					err:   err,
					state: q.state,
				})
				err = nil
				continue
			}
			result = append(result, somaObjectStateResult{
				err:   nil,
				state: state,
			})
		}
	case "show":
		err = r.show_stmt.QueryRow(q.state).Scan(&state)
		if err != nil {
			result = append(result, somaObjectStateResult{
				err:   err,
				state: q.state,
			})
			q.reply <- result
			return
		}

		result = append(result, somaObjectStateResult{
			err:   nil,
			state: state,
		})
	default:
		result = append(result, somaObjectStateResult{
			err:   errors.New("not implemented"),
			state: "",
		})
	}
	q.reply <- result
}

/*
 * Write Access
 */

type somaObjectStateWriteHandler struct {
	input    chan somaObjectStateRequest
	shutdown chan bool
	conn     *sql.DB
	add_stmt *sql.Stmt
	del_stmt *sql.Stmt
	ren_stmt *sql.Stmt
	appLog   *log.Logger
	reqLog   *log.Logger
	errLog   *log.Logger
}

func (w *somaObjectStateWriteHandler) run() {
	var err error

	for statement, prepStmt := range map[string]*sql.Stmt{
		stmt.ObjectStateAdd:    w.add_stmt,
		stmt.ObjectStateDel:    w.del_stmt,
		stmt.ObjectStateRename: w.ren_stmt,
	} {
		if prepStmt, err = w.conn.Prepare(statement); err != nil {
			w.errLog.Fatal(`object_state`, err, stmt.Name(statement))
		}
		defer prepStmt.Close()
	}

	for {
		select {
		case <-w.shutdown:
			break
		case req := <-w.input:
			w.process(&req)
		}
	}
}

func (w *somaObjectStateWriteHandler) process(q *somaObjectStateRequest) {
	var res sql.Result
	var err error

	result := make([]somaObjectStateResult, 0)
	switch q.action {
	case "add":
		res, err = w.add_stmt.Exec(q.state)
	case "delete":
		res, err = w.del_stmt.Exec(q.state)
	case "rename":
		res, err = w.ren_stmt.Exec(q.rename, q.state)
	default:
		result = append(result, somaObjectStateResult{
			err:   errors.New("not implemented"),
			state: "",
		})
		q.reply <- result
		return
	}
	if err != nil {
		result = append(result, somaObjectStateResult{
			err:   err,
			state: q.state,
		})
		q.reply <- result
		return
	}

	rowCnt, _ := res.RowsAffected()
	if rowCnt == 0 {
		result = append(result, somaObjectStateResult{
			err:   errors.New("No rows affected"),
			state: q.state,
		})
	} else if rowCnt > 1 {
		result = append(result, somaObjectStateResult{
			err:   errors.New("Too many rows affected"),
			state: q.state,
		})
	} else {
		result = append(result, somaObjectStateResult{
			err:   nil,
			state: q.state,
		})
	}
	q.reply <- result
}

/* Ops Access
 */
func (r *somaObjectStateReadHandler) shutdownNow() {
	r.shutdown <- true
}

func (w *somaObjectStateWriteHandler) shutdownNow() {
	w.shutdown <- true
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
