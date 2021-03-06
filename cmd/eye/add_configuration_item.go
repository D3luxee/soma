/*
Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
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
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mjolnir42/soma/lib/proto"

	"github.com/julienschmidt/httprouter"
)

func AddConfigurationItem(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		dec      *json.Decoder
		item     *ConfigurationItem
		lookupID string
		raw      *proto.Result
		details  *proto.Deployment
		err      error
	)

	dec = json.NewDecoder(r.Body)
	if err = dec.Decode(raw); err != nil {
		dispatchUnprocessable(&w, err.Error())
		return
	}

	if len(*raw.Deployments) != 1 {
		dispatchBadRequest(&w, "Invalid number of deployments in requests")
		return
	}
	*details = (*raw.Deployments)[0]

	if details.Task != "rollout" {
		dispatchBadRequest(&w, "Deprovision request against rollout API")
		return
	}

	if lookupID, item, err = Itemize(details); err != nil {
		dispatchUnprocessable(&w, err.Error())
		return
	}

	if err = addItem(item, lookupID); err != nil {
		dispatchInternalServerError(&w, err.Error())
		return
	}
	dispatchNoContent(&w)
}

func addItem(item *ConfigurationItem, lookupID string) error {
	var (
		hostid int64
		err    error
		look   string
		jsonb  []byte
	)

	// string was generated from uint64, we need int now
	if hostid, err = strconv.ParseInt(item.HostID, 10, 64); err != nil {
		return err
	}
	if jsonb, err = json.Marshal(item); err != nil {
		return err
	}

lookupcheck:
	err = Eye.run.checkLookup.QueryRow(lookupID).Scan(&look)
	if err == sql.ErrNoRows {
		if _, err = Eye.run.insertLookup.Exec(
			lookupID,
			int(hostid),
			item.Metric,
		); err != nil {
			return err
		}
		goto lookupcheck
	} else if err != nil {
		return err
	}
	if lookupID != look {
		panic(`Database corrupted`)
	}

	_, err = Eye.run.insertItem.Exec(
		item.ConfigurationItemID.String(),
		lookupID,
		jsonb,
	)
	return err
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
