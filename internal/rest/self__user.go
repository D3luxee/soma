/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016-2018, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/soma/internal/rest"

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"
)

// UserShow function
func (x *Rest) UserShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionUser
	request.Action = msg.ActionShow
	request.User.ID = params.ByName(`userID`)

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// UserSearch function
func (x *Rest) UserSearch(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	cReq := proto.NewUserFilter()
	if err := decodeJSONBody(r, &cReq); err != nil {
		dispatchBadRequest(&w, err)
		return
	}

	request := newRequest(r, params)
	request.Section = msg.SectionUser
	request.Action = msg.ActionSearch
	request.Search.User.UserName = cReq.Filter.User.UserName

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
