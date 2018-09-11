/*-
 * Copyright (c) 2016-2018, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/soma/internal/rest"

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"
)

// RepositoryMgmtCreate function
func (x *Rest) RepositoryMgmtCreate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	cReq := proto.NewRepositoryRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		dispatchBadRequest(&w, err)
		return
	}

	nameLen := utf8.RuneCountInString(cReq.Repository.Name)
	if nameLen < 4 || nameLen > 128 {
		dispatchBadRequest(&w, fmt.Errorf(`Illegal repository name length (4 < x <= 128)`))
		return
	}

	request := newRequest(r, params)
	request.Section = msg.SectionRepositoryMgmt
	request.Action = msg.ActionCreate
	request.Repository = cReq.Repository.Clone()

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	send(&w, &result)
}

// RepositoryMgmtRename function
func (x *Rest) RepositoryMgmtRename(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	cReq := proto.NewRepositoryRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		dispatchBadRequest(&w, err)
		return
	}

	nameLen := utf8.RuneCountInString(cReq.Repository.Name)
	if nameLen < 4 || nameLen > 128 {
		dispatchBadRequest(&w, fmt.Errorf(`Illegal new repository name length (4 < x <= 128)`))
		return
	}

	request := newRequest(r, params)
	request.Section = msg.SectionRepositoryMgmt
	request.Action = msg.ActionRename
	request.Repository.ID = params.ByName(`repositoryID`)
	request.Update.Repository.Name = cReq.Repository.Name

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	send(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix