/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package super // import "github.com/mjolnir42/soma/internal/super"

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/proto"
	uuid "github.com/satori/go.uuid"
)

func (s *Supervisor) section(q *msg.Request) {
	result := msg.FromRequest(q)

	s.requestLog(q)

	switch q.Action {
	case msg.ActionList,
		msg.ActionShow,
		msg.ActionSearch:
		go func() { s.sectionRead(q) }()
	case msg.ActionAdd,
		msg.ActionRemove:
		if s.readonly {
			result.Conflict(fmt.Errorf(`Readonly instance`))
			goto abort
		}
		s.sectionWrite(q)
	default:
		result.UnknownRequest(q)
		goto abort
	}
	return

abort:
	q.Reply <- result
}

func (s *Supervisor) sectionRead(q *msg.Request) {
	result := msg.FromRequest(q)

	switch q.Action {
	case msg.ActionList:
		s.sectionList(q, &result)
	case msg.ActionShow:
		s.sectionShow(q, &result)
	case msg.ActionSearch:
		s.sectionSearch(q, &result)
	}

	q.Reply <- result
}

func (s *Supervisor) sectionList(q *msg.Request, r *msg.Result) {
	r.SectionObj = []proto.Section{}
	var (
		err                    error
		rows                   *sql.Rows
		sectionID, sectionName string
	)

	if _, err = s.stmtSectionList.Query(); err != nil {
		r.ServerError(err, q.Section)
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&sectionID,
			&sectionName,
		); err != nil {
			r.ServerError(err, q.Section)
			return
		}
		r.SectionObj = append(r.SectionObj, proto.Section{
			Id:   sectionID,
			Name: sectionName,
		})
	}
	if err = rows.Err(); err != nil {
		r.ServerError(err, q.Section)
	}
}

func (s *Supervisor) sectionShow(q *msg.Request, r *msg.Result) {
	var (
		err                                    error
		sectionID, sectionName, category, user string
		ts                                     time.Time
	)

	if err = s.stmtSectionShow.QueryRow(q.SectionObj.Id).Scan(
		&sectionID,
		&sectionName,
		&category,
		&user,
		&ts,
	); err == sql.ErrNoRows {
		r.NotFound(err, q.Section)
		return
	} else if err != nil {
		r.ServerError(err, q.Section)
		return
	}
	r.SectionObj = []proto.Section{proto.Section{
		Id:       sectionID,
		Name:     sectionName,
		Category: category,
		Details: &proto.DetailsCreation{
			CreatedAt: ts.Format(msg.RFC3339Milli),
			CreatedBy: user,
		},
	}}
}

func (s *Supervisor) sectionSearch(q *msg.Request, r *msg.Result) {
	r.SectionObj = []proto.Section{}
	var (
		err                    error
		rows                   *sql.Rows
		sectionID, sectionName string
	)

	if _, err = s.stmtSectionSearch.Query(
		q.SectionObj.Name); err != nil {
		r.ServerError(err, q.Section)
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&sectionID,
			&sectionName,
		); err != nil {
			r.ServerError(err, q.Section)
			return
		}
		r.SectionObj = append(r.SectionObj, proto.Section{
			Id:   sectionID,
			Name: sectionName,
		})
	}
	if err = rows.Err(); err != nil {
		r.ServerError(err, q.Section)
	}
}

func (s *Supervisor) sectionWrite(q *msg.Request) {
	result := msg.FromRequest(q)

	switch q.Action {
	case msg.ActionAdd:
		s.sectionAdd(q, &result)
	case msg.ActionRemove:
		s.sectionRemove(q, &result)
	}

	if result.IsOK() {
		s.Update <- msg.CacheUpdateFromRequest(q)
	}

	q.Reply <- result
}

func (s *Supervisor) sectionAdd(q *msg.Request, r *msg.Result) {
	var (
		err error
		res sql.Result
	)
	q.SectionObj.Id = uuid.NewV4().String()
	if res, err = s.stmtSectionAdd.Exec(
		q.SectionObj.Id,
		q.SectionObj.Name,
		q.SectionObj.Category,
		q.AuthUser,
	); err != nil {
		r.ServerError(err, q.Section)
		return
	}
	if r.RowCnt(res.RowsAffected()) {
		r.SectionObj = []proto.Section{q.SectionObj}
	}
}

func (s *Supervisor) sectionRemove(q *msg.Request, r *msg.Result) {
	var (
		err error
		tx  *sql.Tx
		res sql.Result
	)
	txMap := map[string]*sql.Stmt{}

	// open multi-statement transaction
	if tx, err = s.conn.Begin(); err != nil {
		r.ServerError(err, q.Section)
		return
	}

	// prepare statements for this transaction
	for name, statement := range map[string]string{
		`action_tx_remove`:     stmt.ActionRemove,
		`action_tx_removeMap`:  stmt.ActionRemoveFromMap,
		`section_tx_remove`:    stmt.SectionRemove,
		`section_tx_removeMap`: stmt.SectionRemoveFromMap,
		`section_tx_actlist`:   stmt.SectionListActions,
	} {
		if txMap[name], err = tx.Prepare(statement); err != nil {
			err = fmt.Errorf("s.SectionTx.Prepare(%s) error: %s",
				name, err.Error())
			r.ServerError(err, q.Section)
			tx.Rollback()
			return
		}
	}

	if res, err = s.sectionRemoveTx(q.SectionObj.Id,
		txMap); err != nil {
		r.ServerError(err, q.Section)
		tx.Rollback()
		return
	}
	// sets r.OK()
	if !r.RowCnt(res.RowsAffected()) {
		tx.Rollback()
		return
	}

	// close transaction
	if err = tx.Commit(); err != nil {
		r.ServerError(err, q.Section)
		return
	}

	r.ActionObj = []proto.Action{q.ActionObj}
}

func (s *Supervisor) sectionRemoveTx(id string,
	txMap map[string]*sql.Stmt) (sql.Result, error) {
	var (
		err      error
		res      sql.Result
		rows     *sql.Rows
		actionID string
		affected int64
	)

	// remove all actions in this section
	if rows, err = txMap[`section_tx_actlist`].Query(
		id); err != nil {
		return res, err
	}

	for rows.Next() {
		if err = rows.Scan(
			&actionID,
		); err != nil {
			rows.Close()
			return res, err
		}
		if res, err = s.actionRemoveTx(actionID, txMap); err != nil {
			rows.Close()
			return res, err
		}
		if affected, err = res.RowsAffected(); err != nil {
			rows.Close()
			return res, err
		} else if affected != 1 {
			rows.Close()
			return res, fmt.Errorf("Delete statement caught %d rows"+
				" of actions instead of 1 (actionID=%s)", affected,
				actionID)
		}
	}
	if err = rows.Err(); err != nil {
		return res, err
	}

	// remove section from all permissions
	if res, err = txMap[`section_tx_removeMap`].Exec(id); err != nil {
		return res, err
	}

	// remove section
	return txMap[`section_tx_remove`].Exec(id)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix