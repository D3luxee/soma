/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"
)

// WorkflowSummary returns information about the current workflow
// status distribution
func WorkflowSummary(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `workflow`,
		Action:     `summary`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`workflow_r`].(*workflowRead)
	handler.input <- msg.Request{
		Section:    `workflow`,
		Action:     `summary`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// WorkflowList function
func WorkflowList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `workflow`,
		Action:     `list`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	cReq := proto.NewWorkflowFilter()
	if err := DecodeJsonBody(r, &cReq); err != nil {
		DispatchBadRequest(&w, err)
		return
	}
	if cReq.Filter.Workflow.Status == `` {
		DispatchBadRequest(&w, fmt.Errorf(
			`No workflow status specified`))
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`workflow_r`].(*workflowRead)
	handler.input <- msg.Request{
		Section:    `workflow`,
		Action:     `list`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Workflow: proto.Workflow{
			Status: cReq.Filter.Workflow.Status,
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// WorkflowRetry function
func WorkflowRetry(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `workflow`,
		Action:     `retry`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	cReq := proto.NewWorkflowRequest()
	if err := DecodeJsonBody(r, &cReq); err != nil {
		DispatchBadRequest(&w, err)
		return
	}
	if cReq.Workflow.InstanceId == `` {
		DispatchBadRequest(&w, fmt.Errorf(
			`No instanceID for retry specified`))
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`workflow_w`].(*workflowWrite)
	handler.input <- msg.Request{
		Section:    `workflow`,
		Action:     `retry`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Workflow: proto.Workflow{
			InstanceId: cReq.Workflow.InstanceId,
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// WorkflowSet function
func WorkflowSet(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `workflow`,
		Action:     `set`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	cReq := proto.NewWorkflowRequest()
	if err := DecodeJsonBody(r, &cReq); err != nil {
		DispatchBadRequest(&w, err)
		return
	}
	if cReq.Workflow.Status == `` || cReq.Workflow.NextStatus == `` ||
		params.ByName(`instanceconfig`) == `` {
		DispatchBadRequest(&w, fmt.Errorf(
			`Incomplete status information specified`))
		return
	}
	// It's dangerous out there, take this -f
	if !cReq.Flags.Forced {
		DispatchBadRequest(&w, fmt.Errorf(
			`WorkflowSet request declined, force required.`))
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`workflow_w`].(*workflowWrite)
	handler.input <- msg.Request{
		Section:    `workflow`,
		Action:     `set`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Workflow: proto.Workflow{
			InstanceConfigId: params.ByName(`instanceconfig`),
			Status:           cReq.Workflow.Status,
			NextStatus:       cReq.Workflow.NextStatus,
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
