package main

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

/*Read functions
 */
func ListMode(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	defer PanicCatcher(w)

	returnChannel := make(chan somaResult)
	handler := handlerMap["modeReadHandler"].(somaModeReadHandler)
	handler.input <- somaModeRequest{
		action: "list",
		reply:  returnChannel,
	}
	result := <-returnChannel
	SendModeReply(&w, &result)
}

func ShowMode(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	returnChannel := make(chan somaResult)
	handler := handlerMap["modeReadHandler"].(somaModeReadHandler)
	handler.input <- somaModeRequest{
		action: "show",
		reply:  returnChannel,
		Mode: somaproto.ProtoMode{
			Mode: params.ByName("mode"),
		},
	}
	result := <-returnChannel
	SendModeReply(&w, &result)
}

/* Write functions
 */
func AddMode(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	defer PanicCatcher(w)

	cReq := somaproto.ProtoRequestMode{}
	err := DecodeJsonBody(r, &cReq)
	if err != nil {
		DispatchBadRequest(&w, err)
		return
	}

	returnChannel := make(chan somaResult)
	handler := handlerMap["modeWriteHandler"].(somaModeWriteHandler)
	handler.input <- somaModeRequest{
		action: "add",
		reply:  returnChannel,
		Mode: somaproto.ProtoMode{
			Mode: cReq.Mode.Mode,
		},
	}
	result := <-returnChannel
	SendModeReply(&w, &result)
}

func DeleteMode(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	returnChannel := make(chan somaResult)
	handler := handlerMap["modeWriteHandler"].(somaModeWriteHandler)
	handler.input <- somaModeRequest{
		action: "delete",
		reply:  returnChannel,
		Mode: somaproto.ProtoMode{
			Mode: params.ByName("mode"),
		},
	}
	result := <-returnChannel
	SendModeReply(&w, &result)
}

/* Utility
 */
func SendModeReply(w *http.ResponseWriter, r *somaResult) {
	result := somaproto.ProtoResultMode{}
	if r.MarkErrors(&result) {
		goto dispatch
	}
	result.Text = make([]string, 0)
	result.Modes = make([]somaproto.ProtoMode, 0)
	for _, i := range (*r).Modes {
		result.Modes = append(result.Modes, i.Mode)
		if i.ResultError != nil {
			result.Text = append(result.Text, i.ResultError.Error())
		}
	}

dispatch:
	json, err := json.Marshal(result)
	if err != nil {
		DispatchInternalError(w, err)
		return
	}
	DispatchJsonReply(w, &json)
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix