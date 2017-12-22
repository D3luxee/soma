/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package rest // import "github.com/mjolnir42/soma/internal/rest"

import (
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/auth"
	"github.com/satori/go.uuid"
)

// SupervisorValidate is a noop function wrapped in HTTP basic
// authentication that can be used to verify one's credentials
func (x *Rest) SupervisorValidate(w http.ResponseWriter, _ *http.Request,
	_ httprouter.Params) {
	w.WriteHeader(http.StatusNoContent)
}

// SupervisorKex is used by the client to initiate a key exchange
// that can the be used for one of the encrypted endpoints
func (x *Rest) SupervisorKex(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	kex := auth.Kex{}
	if err := decodeJSONBody(r, &kex); err != nil {
		dispatchBadRequest(&w, err)
		return
	}

	// generate RequestID
	params = append(params, httprouter.Param{
		Key:   `RequestID`,
		Value: uuid.NewV4().String(),
	})

	request := newRequest(r, params)
	request.Section = msg.SectionSupervisor
	request.Action = msg.ActionKex
	request.Super = &msg.Supervisor{
		Kex: auth.Kex{
			Public:               kex.Public,
			InitializationVector: kex.InitializationVector,
		},
	}

	handler := x.handlerMap.Get(`supervisor`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// SupervisorTokenInvalidate is the rest endpoint to invalidate
// the current access token used during BasicAuth
func (x *Rest) SupervisorTokenInvalidate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionSupervisor
	request.Action = msg.ActionToken
	request.Super = &msg.Supervisor{
		Task:      msg.TaskInvalidate,
		AuthToken: params.ByName(`AuthenticatedToken`),
	}

	// authorization to invalidate the token is implicit from being
	// able to use it for BasicAuth authentication

	handler := x.handlerMap.Get(`supervisor`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// SupervisorTokenInvalidateSelf is the rest endpoint for all users
// to invalidate all current access tokens of theirselves
func (x *Rest) SupervisorTokenInvalidateSelf(w http.ResponseWriter,
	r *http.Request, params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionSupervisor
	request.Action = msg.ActionRevokeTokens
	request.Super = &msg.Supervisor{
		Task:            msg.TaskInvalidateAccount,
		RevokeTokensFor: params.ByName(`AuthenticatedUser`),
	}

	// authorization to invalidate all tokens is implicit from being
	// able to authenticate with this account

	handler := x.handlerMap.Get(`supervisor`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// SupervisorTokenInvalidateAccount is the rest endpoint for admins
// to invalidate all current access tokens for another user
func (x *Rest) SupervisorTokenInvalidateAccount(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionSystemOperation
	request.Action = msg.ActionRevokeTokens
	request.Super = &msg.Supervisor{
		Task:            msg.TaskInvalidateAccount,
		RevokeTokensFor: params.ByName(`account`),
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	handler := x.handlerMap.Get(`supervisor`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// SupervisorTokenInvalidateGlobal is the rest endpoint for admins
// to invalidate all current access tokens
func (x *Rest) SupervisorTokenInvalidateGlobal(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	request := newRequest(r, params)
	request.Section = msg.SectionSystemOperation
	request.Action = msg.ActionRevokeTokens
	request.Super = &msg.Supervisor{
		Task: msg.TaskInvalidateGlobal,
	}

	if !x.isAuthorized(&request) {
		dispatchForbidden(&w, nil)
		return
	}

	handler := x.handlerMap.Get(`supervisor`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(&w, &result)
}

// SupervisorTokenRequest is the encrypted endpoint used to
// request a password token
func (x *Rest) SupervisorTokenRequest(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	x.SupervisorEncryptedData(&w, r, &params, `token/request`)
}

// SupervisorActivateUser is the encrypted endpoint used to
// activate a user account using external ownership verification
func (x *Rest) SupervisorActivateUser(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	x.SupervisorEncryptedData(&w, r, &params, `activate/user`)
}

// SupervisorActivateRoot is the encrypted endpoint used to
// activate the root account using external ownership verification
func (x *Rest) SupervisorActivateRoot(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	x.SupervisorEncryptedData(&w, r, &params, `activate/root`)
}

// SupervisorPasswordChange is the encrypted endpoint used
// to change the account password using the current one.
func (x *Rest) SupervisorPasswordChange(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	x.SupervisorEncryptedData(&w, r, &params, `password/change`)
}

// SupervisorPasswordReset is the encrypted endpoint used to change the account
// password using external ownership verification
func (x *Rest) SupervisorPasswordReset(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer panicCatcher(w)

	x.SupervisorEncryptedData(&w, r, &params, `password/reset`)
}

// SupervisorEncryptedData is the generic function for
// encrypted endpoints
func (x *Rest) SupervisorEncryptedData(w *http.ResponseWriter,
	r *http.Request, params *httprouter.Params, reqType string) {
	defer panicCatcher(*w)

	data := make([]byte, r.ContentLength)
	io.ReadFull(r.Body, data)

	var action, task string
	section := msg.SectionSupervisor
	switch reqType {
	case `token/request`:
		action = msg.ActionToken
		task = msg.TaskRequest
	case `password/reset`:
		action = msg.ActionPassword
		task = msg.TaskReset
	case `password/change`:
		action = msg.ActionPassword
		task = msg.TaskChange
	case `activate/user`:
		action = msg.ActionActivate
		task = msg.SubjectUser
	case `activate/root`:
		action = msg.ActionActivate
		task = msg.SubjectRoot
	}

	request := newRequest(r, *params)
	request.Section = section
	request.Action = action
	request.Super = &msg.Supervisor{
		RestrictedEndpoint: x.restricted,
		Task:               task,
		Encrypted: struct {
			KexID string
			Data  []byte
		}{
			KexID: params.ByName(`kexID`),
			Data:  data,
		},
	}

	handler := x.handlerMap.Get(`supervisor`)
	handler.Intake() <- request
	result := <-request.Reply
	sendMsgResult(w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix