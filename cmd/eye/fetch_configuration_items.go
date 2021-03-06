/*
 * Copyright (c) 2016,2018, 1&1 Internet SE
 * Written by Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved.
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-resty/resty"
	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/lib/proto"
)

type NotifyMessage struct {
	UUID string `json:"uuid" valid:"uuidv4"`
	Path string `json:"path" valid:"abspath"`
}

func FetchConfigurationItems(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		dec    *json.Decoder
		msg    NotifyMessage
		err    error
		soma   *url.URL
		client *resty.Client
		resp   *resty.Response
		res    proto.Result
		ok     bool
	)
	dec = json.NewDecoder(r.Body)
	if err = dec.Decode(&msg); err != nil {
		log.Printf("Received bad notify event: %s\n", err.Error())
		dispatchBadRequest(&w, err.Error())
		return
	}
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.TagMap["abspath"] = govalidator.Validator(func(str string) bool {
		return filepath.IsAbs(str)
	})
	if ok, err = govalidator.ValidateStruct(msg); !ok {
		log.Printf("Failed notify event verification: %s\n", err.Error())
		dispatchBadRequest(&w, err.Error())
		return
	}

	soma, _ = url.Parse(Eye.Soma.url.String())
	soma.Path = strings.Replace(fmt.Sprintf("%s/%s", msg.Path, msg.UUID), `//`, `/`, -1)
	client = resty.New().SetTimeout(500 * time.Millisecond)
	log.Printf("Fetching deployment: %s\n", soma.String())
	if resp, err = client.R().Get(soma.String()); err != nil || resp.StatusCode() > 299 {
		if err == nil {
			err = fmt.Errorf(resp.Status())
		}
		log.Printf("Failed to fetch deployment from SOMA: %s\n", err.Error())
		dispatchPrecondition(&w, err.Error())
		Failed(msg.UUID)
		return
	}
	if err = json.Unmarshal(resp.Body(), &res); err != nil {
		log.Printf("Error deserializing deployment: %s\n", err.Error())
		dispatchUnprocessable(&w, err.Error())
		Failed(msg.UUID)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("Error in fetched deployment, Statuscode %d\n", res.StatusCode)
		dispatchGone(&w, err.Error())
		Failed(msg.UUID)
		return
	}
	if len(*res.Deployments) != 1 {
		log.Printf("Error, deployment contained wrong deployment count: %d\n", len(*res.Deployments))
		dispatchPrecondition(&w, err.Error())
		Failed(msg.UUID)
		return
	}
	if err = CheckUpdateOrInsertOrDelete(&(*res.Deployments)[0]); err != nil {
		log.Printf("Error processing fetched deployment: %s\n", err.Error())
		dispatchInternalServerError(&w, err.Error())
		Failed(msg.UUID)
		return
	}
	dispatchNoContent(&w)
	Success(msg.UUID)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
