/*-
Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"
)

// OutputTree function
func OutputTree(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if params.ByName(`tree`) != `tree` {
		DispatchBadRequest(&w, nil)
		return
	}

	treeT := ``
	switch {
	case params.ByName(`repository`) != ``:
		treeT = `repository`
	case params.ByName(`bucket`) != ``:
		treeT = `bucket`
	case params.ByName(`group`) != ``:
		treeT = `group`
	case params.ByName(`cluster`) != ``:
		treeT = `cluster`
	default:
		DispatchBadRequest(&w, nil)
		return
	}

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:     params.ByName(`AuthenticatedUser`),
		RemoteAddr:   extractAddress(r.RemoteAddr),
		Section:      treeT,
		Action:       `tree`,
		RepositoryID: params.ByName(`repository`),
		BucketID:     params.ByName(`bucket`),
		GroupID:      params.ByName(`group`),
		ClusterID:    params.ByName(`cluster`),
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`tree_r`].(*outputTree)
	handler.input <- msg.Request{
		Section:    `tree`,
		Action:     `output_tree`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Tree: proto.Tree{
			ID:   params.ByName(treeT),
			Type: treeT,
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
