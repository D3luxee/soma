/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/soma/internal/rest"

import (
	"fmt"
	"net/http"

	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"

	"github.com/julienschmidt/httprouter"
)

// CheckConfigList function
func (x *Rest) CheckConfigList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionCheckConfig
	request.Action = msg.ActionList
	request.CheckConfig = proto.CheckConfig{
		RepositoryID: params.ByName(`repositoryID`),
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	send(&w, &result)
}

// CheckConfigSearch function
func (x *Rest) CheckConfigSearch(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	cReq := proto.NewCheckConfigRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		dispatchBadRequest(&w, err)
		return
	}

	if cReq.Filter.CheckConfig.Name == `` {
		dispatchBadRequest(&w, fmt.Errorf(`CheckConfigSearch on empty name`))
		return
	}

	request := newRequest(r, params)
	request.Section = msg.SectionCheckConfig
	request.Action = msg.ActionSearch
	request.CheckConfig = proto.CheckConfig{
		RepositoryID: params.ByName(`repositoryID`),
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply

	filtered := []proto.CheckConfig{}
	for _, i := range result.CheckConfig {
		if i.Name == cReq.Filter.CheckConfig.Name {
			filtered = append(filtered, i)
		}
	}
	result.CheckConfig = filtered
	send(&w, &result)
}

// CheckConfigShow function
func (x *Rest) CheckConfigShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionCheckConfig
	request.Action = msg.ActionShow
	request.CheckConfig = proto.CheckConfig{
		ID:           params.ByName(`checkID`),
		RepositoryID: params.ByName(`repositoryID`),
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	send(&w, &result)
}

// CheckConfigCreate function
func (x *Rest) CheckConfigCreate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	cReq := proto.NewCheckConfigRequest()
	if err := decodeJSONBody(r, &cReq); err != nil {
		dispatchBadRequest(&w, err)
		return
	}

	request := newRequest(r, params)
	request.Section = msg.SectionMonitoring
	request.Action = msg.ActionUse
	request.CheckConfig = cReq.CheckConfig.Clone()

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	request.Section = msg.SectionCheckConfig
	request.Action = msg.ActionCreate

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	send(&w, &result)
}

// CheckConfigDestroy function
func (x *Rest) CheckConfigDestroy(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionMonitoring
	request.Action = msg.ActionUse
	request.CheckConfig = proto.CheckConfig{
		ID:           params.ByName(`checkID`),
		RepositoryID: params.ByName(`repositoryID`),
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	request.Section = msg.SectionCheckConfig
	request.Action = msg.ActionDestroy

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	x.handlerMap.MustLookup(&request).Intake() <- request
	result := <-request.Reply
	send(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix