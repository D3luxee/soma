/*-
 * Copyright (c) 2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package rest implements the REST routes to access SOMA.
package rest // import "github.com/1and1/soma/internal/rest"

import (
	"github.com/1and1/soma/internal/msg"
	"github.com/1and1/soma/internal/soma"
)

// Rest holds the required state for the REST interface
type Rest struct {
	isAuthorized func(*msg.Authorization) bool
	handlerMap   *soma.HandlerMap
}

// New returns a new REST interface
func New(
	authorizationFunction func(*msg.Authorization) bool,
	appHandlerMap *soma.HandlerMap,
) *Rest {
	r := Rest{}
	r.isAuthorized = authorizationFunction
	r.handlerMap = appHandlerMap
	return &r
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix